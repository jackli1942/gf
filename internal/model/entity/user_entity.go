package entity

import "github.com/gogf/gf/v2/os/gtime"

type User struct {
	Id           uint64      `json:"id"           orm:"id"`
	Username     string      `json:"username"     orm:"username"`
	PasswordHash string      `json:"-"            orm:"password_hash"` // Ensure ORM includes it
	Nickname     string      `json:"nickname"     orm:"nickname"`
	Email        string      `json:"email"        orm:"email"`
	Status       int         `json:"status"       orm:"status"`
	CreatedAt    *gtime.Time `json:"createdAt"    orm:"created_at"`
	UpdatedAt    *gtime.Time `json:"updatedAt"    orm:"updated_at"`
}
