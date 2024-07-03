package ratelimit

import (
	"basic-go/webook/internal/service/sms"
	smsmocks "basic-go/webook/internal/service/sms/mocks"
	"basic-go/webook/pkg/ratelimit"
	limitmocks "basic-go/webook/pkg/ratelimit/mocks"
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestRatelimiitSMSServiceSend(t *testing.T) {
	testCases := []struct {
		name string
		mock func(ctrl *gomock.Controller) (sms.Service, ratelimit.Limiter)

		// 测试限流,输入什么并不关键,所以不需要

		wantErr error
	}{
		{
			name: "正常发送",
			mock: func(ctrl *gomock.Controller) (sms.Service, ratelimit.Limiter) {
				// limiter 验证没有限流 true 后发送 Send
				svc := smsmocks.NewMockService(ctrl)
				limiter := limitmocks.NewMockLimiter(ctrl)
				limiter.EXPECT().Limit(gomock.Any(), gomock.Any()).Return(false, nil)
				svc.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
				return svc, limiter
			},
			wantErr: nil,
		},
		{
			name: "触发限流",
			mock: func(ctrl *gomock.Controller) (sms.Service, ratelimit.Limiter) {
				svc := smsmocks.NewMockService(ctrl)
				limiter := limitmocks.NewMockLimiter(ctrl)
				// false 限流后不再触发 Send
				limiter.EXPECT().Limit(gomock.Any(), gomock.Any()).Return(true, nil)
				return svc, limiter
			},
			wantErr: errors.New("短信服务触发了限流"),
		},
		{
			name: "系统错误",
			mock: func(ctrl *gomock.Controller) (sms.Service, ratelimit.Limiter) {
				svc := smsmocks.NewMockService(ctrl)
				limiter := limitmocks.NewMockLimiter(ctrl)
				limiter.EXPECT().Limit(gomock.Any(), gomock.Any()).Return(false, errors.New("系统错误"))
				return svc, limiter
			},
			wantErr: fmt.Errorf("短信服务判断限流出现问题: %w", errors.New("系统错误")),
		},
	}
	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			svc, limiter := tC.mock(ctrl)
			limitSvc := NewRatelimiitSMSService(svc, limiter)
			err := limitSvc.Send(context.Background(), "mytpl", []string{"123"}, "152xxxx")
			assert.Equal(t, tC.wantErr, err)
		})
	}
}
