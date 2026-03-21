package entity

import (
	"time"
)

type User struct {
	UserId			string 		`db:"user_id"`
	Username 		string 		`db:"username"`
	Email 			string 		`db:"email"`
	PasswordHash 	string 		`db:"password_hash"`
	Role			string 		`db:"role"`
	CreatedAt		*time.Time 	`db:"created_at"`
	LastLogin		*time.Time	`db:"last_login"`
	IsActive		bool 		`db:"is_active"`
	Guru			Guru		`db:"guru"`
	
}

