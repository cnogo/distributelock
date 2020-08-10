package distributelock

import (
	"fmt"
	"github.com/go-redis/redis"
	"testing"
	"time"
)

func TestRedisLock_Lock(t *testing.T) {
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
	locker2 := NewRedisLock(cli, key, WithTTL(2 * time.Second))

	fmt.Println(locker1.Lock())
	locker1.UnLock()

	time.Sleep(5 * time.Second)
	fmt.Println(locker2.Lock())
	defer locker2.UnLock()

	time.Sleep(10 * time.Minute)
}
