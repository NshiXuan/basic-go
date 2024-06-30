package ratelimit

import "context"

type Limiter interface {
	// Limited 有没有出发限流，key 是限流对象
	// bool 代表是否限流, ture 就是要限流
	// err 限流器本有没有错误
	Limit(ctx context.Context, key string) (bool, error)
}
