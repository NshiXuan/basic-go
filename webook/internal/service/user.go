package service

import (
	"basic-go/webook/internal/domain"
	"basic-go/webook/internal/repository"
	"context"
	"errors"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserDuplicateEmail = repository.ErrUserDuplicateEmail
	ErrInvalidUserOrEmail = errors.New("账号/邮箱或密码错误")
)

type UserService interface {
	SignUp(ctx context.Context, u domain.User) error
	Login(ctx context.Context, email, password string) (domain.User, error)
	Profile(ctx context.Context, id int64) (domain.User, error)
	FindOrCreate(ctx context.Context, phone string) (domain.User, error)
}

type userService struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) UserService {
	return &userService{
		repo: repo,
	}
}

func (svc *userService) SignUp(ctx context.Context, u domain.User) error {
	// 加密
	hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hash)
	return svc.repo.Create(ctx, u)
}

func (svc *userService) Login(ctx context.Context, email, password string) (domain.User, error) {
	u, err := svc.repo.FindByEmail(ctx, email)
	if err == repository.ErrUserNotFound {
		return domain.User{}, ErrInvalidUserOrEmail
	}
	if err != nil {
		return domain.User{}, err
	}
	// 比较密码
	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password)); err != nil {
		// TODO(nsx): 打日志
		return domain.User{}, ErrInvalidUserOrEmail
	}
	return u, nil
}

func (svc *userService) Profile(ctx context.Context, id int64) (domain.User, error) {
	// 在系统内部 基本上都是使用 ID
	u, err := svc.repo.FindById(ctx, id)
	return u, err
}

func (svc *userService) FindOrCreate(ctx context.Context, phone string) (domain.User, error) {
	// 快路径
	u, err := svc.repo.FindByPhone(ctx, phone)
	if err != repository.ErrUserNotFound {
		// nil 与不为 ErrUserNotFound 都会进来这里
		return u, err
	}
	// 明确知道没有这个用户
	u = domain.User{
		Phone: phone,
	}
	// 慢路径，一旦触发降级，就不要走慢路径（也就是不能创建了）
	// u 是没有 id 的
	err = svc.repo.Create(ctx, u)
	if err != nil && err != repository.ErrUserDuplicateEmail {
		return u, err
	}
	return svc.repo.FindByPhone(ctx, phone)
}
