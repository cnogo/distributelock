package distributelock

import (
	"fmt"
	"github.com/go-redis/redis"
	"testing"
	"time"
)

//func TestEctdLock_Lock(t *testing.T) {
//	cli, err := clientv3.New(clientv3.Config{
//		Endpoints:   []string{"localhost:2379"},
//		DialTimeout: 5 * time.Second,
//	})
//	if err != nil {
//		// handle error!
//	}
//	defer cli.Close()
//}

func aTestRedisLock_Lock(t *testing.T) {
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


	time.Sleep(2 * time.Second)

	fmt.Println(locker2.Lock())

	time.Sleep(2 * time.Second)
	locker1.UnLock()
	defer locker2.UnLock()

	time.Sleep(10 * time.Second)
}
