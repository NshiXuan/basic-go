package ratelimit

import (
	"basic-go/webook/internal/service/sms"
	"basic-go/webook/pkg/ratelimit"
	"context"
	"errors"
	"fmt"
)

var errLimited = errors.New("短信服务触发了限流")

type RatelimiitSMSService struct {
	svc     sms.Service
	limiter ratelimit.Limiter
}

func NewRatelimiitSMSService(svc sms.Service, limiter ratelimit.Limiter) sms.Service {
	return &RatelimiitSMSService{
		svc:     svc,
		limiter: limiter,
	}
}

func (s *RatelimiitSMSService) Send(ctx context.Context, tpl string, args []string, numbers ...string) error {
	limited, err := s.limiter.Limit(ctx, "sms:tencent")
	if err != nil {
		// 系统错误 redis 崩了
		// 可以限流: 保守策略,你的下游很坑的时候
		// 可以不限: 你的下游很强,业务可用性要求很高,尽量容错策略
		return fmt.Errorf("短信服务判断限流出现问题: %w", err)
	}
	if limited {
		return errLimited
	}
	// 这里加新特性
	err = s.svc.Send(ctx, tpl, args, numbers...)
	// 这里也可以加新特性
	return err
}
