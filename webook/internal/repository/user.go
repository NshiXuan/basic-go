package repository

import (
	"basic-go/webook/internal/domain"
	"basic-go/webook/internal/repository/cache"
	"basic-go/webook/internal/repository/dao"
	"context"
)

var (
	ErrUserDuplicateEmail = dao.ErrUserDuplicateEmail
	ErrUserNotFound       = dao.ErrUserNotFound
)

type UserRepository struct {
	dao   *dao.UserDao
	cache *cache.UserCache
}

func NewUserRepository(dao *dao.UserDao, c *cache.UserCache) *UserRepository {
	return &UserRepository{
		dao:   dao,
		cache: c,
	}
}

func (r *UserRepository) Create(ctx context.Context, u domain.User) error {
	return r.dao.Insert(ctx, dao.User{
		Email:    u.Email,
		Password: u.Password,
	})

	// TODO(nsx): 操作缓存
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (domain.User, error) {
	u, err := r.dao.FindByEmail(ctx, email)
	if err != nil {
		return domain.User{}, err
	}
	return domain.User{
		Id:       u.Id,
		Email:    u.Email,
		Password: u.Password,
	}, nil
}

func (r *UserRepository) FindById(ctx context.Context, id int64) (domain.User, error) {
	// 先从 cache 里面找
	u, err := r.cache.Get(ctx, id)
	if err == nil {
		// cache 必然有数据
		return u, nil
	}
	// if err == cache.ErrKeyNotFound {
	//  // cache 里面没有数据，需要从数据库里面找
	// }

	// 这里怎么办？ err = io.EOF
	// 加载: 做好兜底，保护好数据库
	// - 如果 Redis 崩了，10w QPS 会导致 Mysql 崩溃
	//    - 二级缓存（redis、本地等）
	//    - 数据库限流，数据库单机限流
	// - 如果只是一两个请求没拿到，这时候加载没用问题
	// 不加载：用户体验差一点

	// 这里选择加载
	ue, err := r.dao.FindById(ctx, id)
	if err != nil {
		return domain.User{}, err
	}

	u = domain.User{
		Id:       ue.Id,
		Email:    ue.Email,
		Password: ue.Password,
	}
	go func() {
		err = r.cache.Set(ctx, u)
		if err != nil {
			// 缓存设置失败不是多大的问题，打日志，做监控即可
		}
	}()
	return u, nil
}
