package distributelock

import (
	"github.com/go-redis/redis"
	"log"
	"sync/atomic"
	"time"
)

var (
	dval = "1"
)

func newRedisLock(cli *redis.Client, key string, cfg *config) Loker {
	return &RedisLock{
		cnf: cfg,
		key: key,
		redisCli: cli,
		die: make(chan struct{}),
	}
}


// redis锁
type RedisLock struct {
	cnf *config
	redisCli *redis.Client // redis客户端
	key string   // 加锁key值
	retryCnt int32 // 重试此时
	die chan struct{}
}

// 加锁
func (p *RedisLock) Lock() error {

	boolCmd := p.redisCli.SetNX(p.key, dval, p.cnf.TTL)

	if boolCmd.Err() != nil {
		return boolCmd.Err()
	}

	// 锁已经存在
	if !boolCmd.Val() {
		return ErrLockExist
	}

	// 续租协程
	go func() {
		ticker := time.NewTicker(p.cnf.TTL / 3)
		defer ticker.Stop()
		for {
			// ttl/3 续租时间，两次续租机会
			select {
			case <- ticker.C:
				expireBoolCmd := p.redisCli.Expire(p.key, p.cnf.TTL)
				if expireBoolCmd.Err() != nil {
					atomic.AddInt32(&p.retryCnt, 1)
					continue
				}

				// 如果没有键值，不需要续租
				if !expireBoolCmd.Val() {
					log.Println("续租失败，看情况是否需要进行人工处理，key: ", p.key)
					return
				}

				atomic.StoreInt32(&p.retryCnt, 0)

			case <- p.die:
				return
			}
		}
	}()

	return nil
}


// 解锁
func (p *RedisLock) UnLock() {
	close(p.die)
	p.redisCli.Del(p.key)
}
