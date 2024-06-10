package service

import (
	"basic-go/webook/internal/repository"
	"basic-go/webook/internal/service/sms"
	"context"
	"fmt"
	"math/rand"
)

// 需要自己去腾讯云获取
const codeTplId = "1"

var (
	ErrCodeSendTooMany        = repository.ErrCodeSendTooMany
	ErrCodeVerifyTooManyTimes = repository.ErrCodeVerifyTooManyTimes
)

type CodeService interface {
	Send(ctx context.Context, biz string, phone string) error
	Verify(ctx context.Context, biz string, phone string, inputCode string) (bool, error)
}

type codeService struct {
	repo   repository.CodeRepository
	smsSvc sms.Service
}

func NewCodeService(repo repository.CodeRepository, smsSvc sms.Service) CodeService {
	return &codeService{
		repo:   repo,
		smsSvc: smsSvc,
	}
}

func (svc *codeService) Send(ctx context.Context, biz string, phone string) error {
	// 生成验证码
	code := svc.generateCode()
	// 塞进 redis
	err := svc.repo.Store(ctx, biz, phone, code)
	if err != nil {
		return err
	}
	// 发送
	err = svc.smsSvc.Send(ctx, codeTplId, []string{code}, phone)
	// if err != nil {
	//  // 这意味着 Redis 有这个验证码，但这个 err 可能是超时的 err ，不知道有没有发出去
	//  // 可以在这里重试，初始化的时候，传入一个会重试的 smsSvc （需要自己实现）
	//  return err
	// }
	return err
}

// Verify 与 VerifyV1 两种返回值都可以
func (svc *codeService) Verify(ctx context.Context, biz string, phone string, inputCode string) (bool, error) {
	return svc.repo.Verify(ctx, biz, phone, inputCode)
}

// func (c *codeService) VerifyV1(ctx context.Context, biz string, phone string, inputCode string) error {
//  return nil
// }

func (svc *codeService) generateCode() string {
	// 生成包含   0 ~ 999999 的随机数
	// 不够 6 为, %6d 会补上前导 0
	num := rand.Intn(1000000)
	return fmt.Sprintf("%6d", num)
}
