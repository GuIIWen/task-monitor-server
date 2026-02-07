package model

import "time"

// User 用户信息
type User struct {
	ID        uint      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Username  string    `gorm:"column:username;uniqueIndex;size:50;not null" json:"username"`
	Password  string    `gorm:"column:password;size:255;not null" json:"-"`
	CreatedAt time.Time `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updatedAt"`
}

func (User) TableName() string {
	return "users"
}
