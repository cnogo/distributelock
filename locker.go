package distributelock

import "github.com/go-redis/redis"

type Loker interface {
	Lock() error
	UnLock()
}


func NewRedisLock(cli *redis.Client, key string, configers ...configer) Loker {
	cnf := new(config)
	cnf.setDefault()

	for _, configer := range configers {
		configer(cnf)
	}

	return newRedisLock(cli, key, cnf)
}
