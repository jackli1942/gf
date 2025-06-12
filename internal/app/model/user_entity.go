package model

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// User is the an entity representing a user.
type User struct {
	Id        uint        `json:"id"        orm:"id"         description:"User ID"`
	Uuid      string      `json:"uuid"      orm:"uuid"       description:"User UUID"`
	Username  string      `json:"username"  orm:"username"   description:"Username"`
	Password  string      `json:"-"         orm:"password"   description:"Password (hashed)"`
	Nickname  string      `json:"nickname"  orm:"nickname"   description:"Nickname"`
	Email     string      `json:"email"     orm:"email"      description:"Email address"`
	Phone     string      `json:"phone"     orm:"phone"      description:"Phone number"`
	Avatar    string      `json:"avatar"    orm:"avatar"     description:"User avatar URL"`
	Status    int         `json:"status"    orm:"status"     description:"User status (e.g., 1:active, 0:disabled)"`
	CreatedAt *gtime.Time `json:"createdAt" orm:"created_at" description:"Creation time"`
	UpdatedAt *gtime.Time `json:"updatedAt" orm:"updated_at" description:"Last update time"`
}
