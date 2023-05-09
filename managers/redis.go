package managers

import (
	"sync"
	"time"

	"github.com/go-redis/redis/v7"
	"github.com/sirupsen/logrus"
)

var redisLog = logrus.WithField("fun", "db")
var Redis *redis.Client

const (
	//用户缓存生存的时间，超过这一时间用户需要重新登录
	UserCacheLife = 24 * time.Hour
)

func InitRedis(wg *sync.WaitGroup) {
	options := redis.Options{
		Addr:     CONFIG.Redis.URL,
		Password: CONFIG.Redis.Password,
		DB:       CONFIG.Redis.DB, // use default DB
	}

	Redis = redis.NewClient(&options)

	if _, err := Redis.Ping().Result(); err != nil {
		redisLog.WithError(err).Panic("Init redis failed")
	}
	redisLog.Info("Have connected to redis")
	wg.Done()
}

func HMSetAndExpire(key string, fields map[string]interface{}, expiration time.Duration) error {
	if err := Redis.HMSet(key, fields).Err(); err != nil {
		return err
	}
	return Redis.Expire(key, expiration).Err()
}
