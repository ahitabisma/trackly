package user

import "time"

type User struct {
	ID        uint    `gorm:"primaryKey"`
	Name      string  `gorm:"type:varchar(100);not null"`
	Email     string  `gorm:"type:varchar(100);uniqueIndex;not null"`
	Password  string  `gorm:"type:text;not null"`
	Avatar    *string `gorm:"type:text"`
	Role      string  `gorm:"type:varchar(50);default:user"`
	CreatedAt time.Time
	UpdatedAt time.Time
}
