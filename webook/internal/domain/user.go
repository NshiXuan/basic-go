package domain

import "time"

// User 领域对象，是 DDD 的聚合根
// 是 DDD 中的 entity, 有些人也叫 BO (bussion object)
type User struct {
	Id       int64
	Email    string
	Password string
	Phone    string
	Ctime    time.Time
}
