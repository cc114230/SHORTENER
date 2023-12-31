package config

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/rest"
)

type Config struct {
	rest.RestConf

	ShortUrlDB struct { //匿名结构体
		DSN string
	}

	Sequence struct {
		DSN string
	}

	ShortUrlBlackList []string
	ShortDomain       string

	CacheRedis cache.CacheConf // redis缓存
}
