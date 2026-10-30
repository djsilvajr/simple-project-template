package model

import (
	"time"

	"gorm.io/gorm"
)

type Role string

const (
	RoleAdmin Role = "admin"
	RoleUser  Role = "user"
)

func (r Role) IsValid() bool {
	return r == RoleAdmin || r == RoleUser
}

type User struct {
	ID        uint           `gorm:"primaryKey"                               json:"id"`
	Login     string         `gorm:"size:100;not null"                        json:"login"`
	Name      string         `gorm:"size:100;not null"                        json:"name"`
	Email     string         `gorm:"uniqueIndex;size:150;null"                json:"email"`
	Password  string         `gorm:"not null"                                 json:"-"`
	Role      Role           `gorm:"type:enum('admin','user');default:'user'" json:"role"`
	CreatedAt time.Time      `                                                  json:"createdAt"`
	UpdatedAt time.Time      `                                                  json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index"                                      json:"deletedAt,omitempty"`
}
