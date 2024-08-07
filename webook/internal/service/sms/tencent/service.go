package tencent

import (
	"basic-go/webook/pkg/ratelimit"
	"context"
	"fmt"

	"github.com/ecodeclub/ekit/slice"
	sms "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/sms/v20210111"
)

type Service struct {
	appId     *string
	signaName *string
	client    *sms.Client
	limiter   ratelimit.Limiter
}

func NewService(client *sms.Client, appId string, signame string) *Service {
	return &Service{
		client:    client,
		appId:     &appId,
		signaName: &signame,
	}
}

func (s *Service) Send(ctx context.Context, tpl string, args []string, numbers ...string) error {
	req := sms.NewSendSmsRequest()
	req.SmsSdkAppId = s.appId
	req.SignName = s.signaName
	req.TemplateId = &tpl
	req.PhoneNumberSet = s.toStringPtrSlice(numbers)
	req.TemplateParamSet = s.toStringPtrSlice(args)
	resp, err := s.client.SendSms(req)
	if err != nil {
		return err
	}
	for _, status := range resp.Response.SendStatusSet {
		if status.Code != nil && *(status.Code) != "Ok" {
			return fmt.Errorf("发送短信失败: %s, %s", *status.Code, *status.Message)
		}
	}
	return nil
}

func (s *Service) toStringPtrSlice(src []string) []*string {
	return slice.Map[string, *string](src, func(idx int, src string) *string {
		return &src
	})
}
