package cache

import (
	"basic-go/webook/internal/domain"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

var ErrKeyNotFound = redis.Nil

type UserCache struct {
	// Cmdable 接口，传单机的 Redis 可以，传 cluster 的 redis 也可以
	client     redis.Cmdable
	expiration time.Duration
}

// A 用到了 B , B 一定是接口
// A 用到了 B , B 一定是 A 的字段
// A 用到了 B , A 绝对不初始化 B ，而是外面注入
// 不要去搞初始化和包变量
func NewUserCache(client redis.Cmdable) *UserCache {
	return &UserCache{
		client:     client,
		expiration: time.Minute * 15,
	}
}

// 只要 error 为 nil , 就认为缓存有数据
// 如果没有数据，返回一个特定的 error
func (cache *UserCache) Get(ctx context.Context, id int64) (domain.User, error) {
	key := cache.Key(id)
	// 如果数据不存在 err = redis.Nil
	val, err := cache.client.Get(ctx, key).Bytes()
	if err != nil {
		return domain.User{}, err
	}
	var u domain.User
	err = json.Unmarshal(val, &u)
	return u, err
}

func (cache *UserCache) Set(ctx context.Context, u domain.User) error {
	// redis 不能存储自定义的结构体类型，基础类型可以，如果是 protobuf 工具生成可以，redis 会通过 protobuf 序列化
	val, err := json.Marshal(u)
	if err != nil {
		return err
	}
	key := cache.Key(u.Id)
	return cache.client.Set(ctx, key, val, cache.expiration).Err()
}

// 屏蔽 key 的结构,调用者不用知道在缓存里面的这个 key 是怎么组成的
func (cache *UserCache) Key(id int64) string {
	return fmt.Sprintf("user:info:%d", id)
}
