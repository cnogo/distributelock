package distributelock

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"github.com/go-redis/redis"
	"io"
	"log"
	"sync/atomic"
	"time"
)

var (
	// 刷新锁的脚本
	freshLua = redis.NewScript(`
			if redis.call("get", KEYS[1]) == ARGV[1] then
				return redis.call("pexpire", KEYS[1], ARGV[2])
			else
				return 0
			end
		`)

	// 释放锁的脚本
	releaseLua = redis.NewScript(`
			if redis.call("get", KEYS[1]) == ARGV[1] then
				redis.call("del", KEYS[1])
			else
				return 0
			end
		`)
)

func newRedisLock(cli *redis.Client, key string, cfg *config) Locker {
	rl := &RedisLock{
		key: key,
		redisCli: cli,
		die: make(chan struct{}),
	}

	rl.cnf.Store(cfg)

	return rl
}


// redis锁
type RedisLock struct {
	cnf atomic.Value
	redisCli *redis.Client // redis客户端
	key string   // 加锁key值
	retryCnt int32 // 重试此时
	kVal string // 键值
	die chan struct{}
}

// 加锁
func (p *RedisLock) Lock() error {
	cnf := p.cnf.Load().(*config)

	randBytes := make([]byte, 16)

	if _, err := io.ReadFull(rand.Reader, randBytes); err != nil {
		return err
	}

	p.kVal = base64.StdEncoding.EncodeToString(randBytes)


	boolCmd := p.redisCli.SetNX(p.key, p.kVal, cnf.TTL)

	if boolCmd.Err() != nil {
		return boolCmd.Err()
	}

	// 锁已经存在
	if !boolCmd.Val() {
		return ErrLockExist
	}

	// 续租协程
	go func() {
		cnf := p.cnf.Load().(*config)
		ticker := time.NewTicker(cnf.TTL / 3)
		defer ticker.Stop()
		for {
			// ttl/3 续租时间，两次续租机会
			select {
			case <- ticker.C:
				fmt.Println("续租, key ", p.key, " value ", p.kVal)
				resultCmd := freshLua.Run(p.redisCli, []string{p.key}, p.kVal, int64(cnf.TTL/3))
				if resultCmd.Err() != nil {
					log.Println("续租失败重试", resultCmd.Err())
					atomic.AddInt32(&p.retryCnt, 1)
					continue
				}

				if resultCmd.Val() != int64(1) {
					// 分布式锁续租失败，可能业务没有做完，资源就被其他请求抢占，有可能出现脏数据，可能需要进行人工处理
					log.Println("续租失败，打印日志，对业务可能出现脏数据进行人工处理，key: ", p.key, " value: ", p.kVal)
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

	_ = releaseLua.Run(p.redisCli, []string{p.key}, p.kVal)
}
