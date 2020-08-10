package distributelock

import (
	"errors"
	"time"
)

var (
	ErrLockExist = errors.New("lock is exists")
)

type config struct {
	TTL time.Duration  // 生命周期
}

// 设置默认配置
func (p *config) setDefault() {
	p.TTL = 30 * time.Millisecond
}


type configer func(cnf *config)

// 设置锁的生命周期
func WithTTL(ttl time.Duration) configer {
	return func(cnf *config) {
		// 至少得10毫秒
		if ttl < 10 * time.Millisecond {
			ttl = 10 * time.Millisecond
		}
		cnf.TTL = ttl
	}
}
