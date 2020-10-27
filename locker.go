package distributelock

import (
	"github.com/go-redis/redis"
)

type Locker interface {
	Lock() error
	UnLock()
}

// AP场景的分布式锁
func NewRedisLock(cli *redis.Client, key string, configers ...configer) Locker {
	cnf := new(config)
	cnf.setDefault()

	for _, configer := range configers {
		configer(cnf)
	}

	return newRedisLock(cli, key, cnf)
}

