package cache_locker

import (
	"github.com/garyburd/redigo/redis"
	"github.com/ousanki/freya/global"
	"github.com/ousanki/freya/log"
)

type CacheLocker struct {
	RdsName  string
	LockName string
}

func NewCacheLocker(rds string, key string) *CacheLocker {
	return &CacheLocker{
		RdsName:  rds,
		LockName: key,
	}
}

func (cl *CacheLocker) Lock() bool {
	client := global.GetRedis(cl.RdsName)
	defer client.Close()
	ret, err := redis.String(client.Do("SET", cl.LockName, 1, "NX", "EX", 10))

	if err != nil {
		log.GetLogger().Warningf("CacheLocker | Lock | error:%s", err.Error())
		return false
	}
	if ret != "OK" {
		log.GetLogger().Warningf("CacheLocker | Lock | ret:%s", ret)
		return false
	}
	return true
}

func (cl *CacheLocker) UnLock() {
	client := global.GetRedis(cl.RdsName)
	defer client.Close()
	_, err := client.Do("DEL", cl.LockName)
	if err != nil {
		log.GetLogger().Warningf("CacheLocker | UnLock | error:%s", err.Error())
	}
}
