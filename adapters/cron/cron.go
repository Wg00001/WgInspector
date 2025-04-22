package cron

import (
	"WgInspector/entities/config"
	"WgInspector/entities/task"
	"WgInspector/usecase/task/cron"
	"context"
	"fmt"
	"github.com/go-co-op/gocron/v2"
	"github.com/google/uuid"
	"log"
	"strconv"
	"sync"
	"time"
)

/**
 * @description: TODO
 * @author Wg
 * @date 2025/2/14
 */

func init() {
	c := new(Cron)
	c.Init()
	cron.Use(c)
}

type Cron struct {
	s            gocron.Scheduler
	monitorChans map[chan<- []task.Stat]context.Context
	mu           sync.RWMutex
}

var _ task.Cron = (*Cron)(nil)

func (c *Cron) Init() error {
	sTemp, err := gocron.NewScheduler(
		gocron.WithLocation(time.Local),                                           // 设置时区
		gocron.WithGlobalJobOptions(gocron.WithEventListeners(c.afterListener())), // 全局任务选项
	)
	if err != nil {
		return fmt.Errorf("init cron Scheduler fail！: %v", err)
	}
	c.monitorChans = make(map[chan<- []task.Stat]context.Context)
	c.s = sTemp
	return nil
}

func (c *Cron) AddTask(task task.Task) error {
	definition, err := jobDefinition(task.GetCron())
	if err != nil {
		return err
	}
	u := uuid.NewSHA1(uuid.NameSpaceOID, []byte(task.Identity().Name))
	c.s.RemoveJob(u)

	_, err = c.s.NewJob(
		definition,
		gocron.NewTask(func() {
			err := task.Do(context.Background())
			if err != nil {
				log.Printf("gocron do task Err\n- task name: %s\n- err: %v\n--- \n", task.Identity(), err)
				return
			}
		}), // 任务函数和参数

		gocron.WithName(task.Identity().Name),
		gocron.WithIdentifier(u),
	)
	return err
}

func (c *Cron) Start() {
	c.s.Start()
}

func (c *Cron) Exit() {
	err := c.s.StopJobs()
	if err != nil {
		log.Println("cron: " + err.Error())
		return
	}
	log.Println("cron: exit")
}

func (c *Cron) afterListener() gocron.EventListener {
	return gocron.AfterJobRuns(func(uuid.UUID, string) {
		stats, err := c.jobStats()
		if err != nil {
			log.Printf("cron after Listener: %s", err)
			return
		}
		var toDelete []chan<- []task.Stat
		var wg sync.WaitGroup
		c.mu.RLock()
		for ch, ctx := range c.monitorChans {
			wg.Add(1)
			go func(ch chan<- []task.Stat, ctx context.Context) {
				defer wg.Done()
				select {
				case <-ctx.Done():
					toDelete = append(toDelete, ch)
					close(ch)
				default:
					select {
					case ch <- stats:
					default:
					}
				}
			}(ch, ctx)
		}
		wg.Wait()
		c.mu.RUnlock()
		c.mu.Lock()
		defer c.mu.Unlock()
		for _, ch := range toDelete {
			delete(c.monitorChans, ch)
		}
	})
}

// Monitor 调用时传入context，当context.Done时服务端关闭此通道的发送
func (c *Cron) Monitor(ctx context.Context) (<-chan []task.Stat, error) {
	stats, err := c.jobStats()
	if err != nil {
		return nil, err
	}
	resChan := make(chan []task.Stat, 8)
	resChan <- stats
	c.mu.Lock()
	defer c.mu.Unlock()
	c.monitorChans[resChan] = ctx
	return resChan, nil
}

func (c *Cron) jobStats() ([]task.Stat, error) {
	res := make([]task.Stat, 0, len(c.s.Jobs()))
	for _, job := range c.s.Jobs() {
		nextRun, err := job.NextRun()
		if err != nil {
			return nil, err
		}
		lastRun, err := job.LastRun()
		if err != nil {
			return nil, err
		}
		res = append(res, task.Stat{
			UUID:      job.ID().String(),
			TaskName:  job.Name(),
			NextStart: nextRun,
			LastStart: lastRun,
		})
	}
	return res, nil
}

// 将task中的时间设置读取到cron的对象中
func jobDefinition(t config.Cron) (gocron.JobDefinition, error) {
	//if t.Duration == "" {
	//	return nil, fmt.Errorf("gocron add task err, time not define, taskname: %s\n", t)
	//}
	cConfig := t

	//使用cron表达式
	if t.CronTab != "" {
		return gocron.CronJob(t.CronTab, true), nil
	}

	//时间戳周期任务
	if cConfig.Duration != 0 {
		return gocron.DurationJob(cConfig.Duration), nil
	}

	//gocron.OneTimeJob()

	atTime := gocron.NewAtTimes(gocron.NewAtTime(0, 0, 0))
	if len(cConfig.AtTime) != 0 {
		atTime = gocron.NewAtTimes(
			gocron.NewAtTime(parseAtTime(cConfig.AtTime[0])),
			func() []gocron.AtTime {
				res := make([]gocron.AtTime, 0, len(cConfig.AtTime)-1)
				for i := 1; i < len(cConfig.AtTime); i++ {
					res = append(res, gocron.NewAtTime(parseAtTime(cConfig.AtTime[i])))
				}
				return res
			}()...)
	}
	if len(cConfig.Monthly) != 0 {
		return gocron.MonthlyJob(1, func() gocron.DaysOfTheMonth {
			if len(cConfig.Monthly) > 1 {
				return gocron.NewDaysOfTheMonth(cConfig.Monthly[0], cConfig.Monthly[1:]...)
			}
			return gocron.NewDaysOfTheMonth(cConfig.Monthly[0])
		}(), atTime), nil
	}
	if len(cConfig.Weekly) != 0 {
		return gocron.WeeklyJob(1, func() gocron.Weekdays {
			if len(cConfig.Weekly) > 1 {
				return gocron.NewWeekdays(cConfig.Weekly[0], cConfig.Weekly[1:]...)
			}
			return gocron.NewWeekdays(cConfig.Weekly[0])
		}(), atTime), nil
	}
	return gocron.DailyJob(1, atTime), nil
	//return gocron.DurationJob(time.Second * 5)
}

func parseAtTime(t string) (uint, uint, uint) {
	temp := [3]uint64{}
	for l, i := 0, 0; l < len(t) && i < 3; {
		if t[l] < '0' || t[l] > '9' {
			l++
			continue
		}
		r := l + 1
		for r < len(t) && t[r] >= '0' && t[r] <= '9' {
			r++
		}
		temp[i], _ = strconv.ParseUint(t[l:r], 10, 64)
		i++
		l = r
	}
	return uint(temp[0]), uint(temp[1]), uint(temp[2])
}

func (c *Cron) DoNow(key string) error {
	//u := uuid.NewSHA1(uuid.NameSpaceOID, []byte(key))

	for _, job := range c.s.Jobs() {
		if job.ID().String() == key {
			return job.RunNow()
		}
	}
	return fmt.Errorf("task not exist: %s\n", key)
}
