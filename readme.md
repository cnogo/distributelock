## 基于redis的分布式锁实现


示例
```golang
	cli := redis.NewClient(&redis.Options{
		Addr: "120.7.0.1:6739",
		DB: 0,
	})

	_, err := cli.Ping().Result()
	if err != nil {
		t.Error(err)
		return
	}

	key := "123465"
	locker1 := NewRedisLock(cli, key, WithTTL(1 * time.Second))

	fmt.Println(locker1.Lock())
	locker1.UnLock()

    // todo business logic
```