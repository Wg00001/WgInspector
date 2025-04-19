# WgInspector

```text

```

v2：

- 直接使用pgsql存放config
- 用户数据等内容放在pgsql中 
- 去掉配置中心，转而直接操作pgsql，另外添加缓存层
- 所有支持的驱动应该在init时注册，但是具体初始化应该在第一次调用时才完成（惰性初始化）


