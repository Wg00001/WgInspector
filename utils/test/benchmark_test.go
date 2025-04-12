package test

import (
	"WgInspector/adapters/start"
	"context"
	"os"
	"runtime"
	"runtime/pprof"
	"testing"
	"time"
)

func Benchmark(b *testing.B) {
	b.ReportAllocs()
	start.SetLocalConfigReaderOption("../../app/config", "yaml")
	start.InitOld()
	start.Run(context.TODO())
}

func TestPProf(t *testing.T) {
	f, _ := os.Create("mem.prof")
	defer f.Close()
	pprof.WriteHeapProfile(f)
	defer pprof.StopCPUProfile()

	start.SetLocalConfigReaderOption("../../app/config", "yaml")
	start.InitOld()
	ctx, _ := context.WithTimeout(context.TODO(), time.Second*15)
	start.Run(ctx)
}

func TestRuntime(t *testing.T) {
	RuntimeMemStats(t, func() {
		start.SetLocalConfigReaderOption("../../app/config", "yaml")
		start.InitOld()
		start.Run(context.TODO())
	})
}

func RuntimeMemStats(b *testing.T, f func()) {
	var start, end runtime.MemStats

	runtime.GC() // 手动触发GC，减少干扰
	runtime.ReadMemStats(&start)

	f()

	runtime.GC() // 再次触发GC
	runtime.ReadMemStats(&end)

	// 计算并输出内存差异
	b.Logf("堆内存增量: %d bytes", end.HeapAlloc-start.HeapAlloc)
	b.Logf("累计分配量: %d bytes", end.TotalAlloc-start.TotalAlloc)
	b.Logf("内存分配次数: %d", end.Mallocs-start.Mallocs)
}
