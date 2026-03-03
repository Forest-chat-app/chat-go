package model

type Role struct {
	ID          string `gorm:"primaryKey;type:varchar(255)"`
	Name        string `gorm:"type:varchar(255);not null"`
	Description string `gorm:"type:varchar(255);not null"`
	CreatedAt   int    `gorm:"not null"`
	UpdatedAt   int    `gorm:"not null"`
	RoleId      int16  `gorm:"type:int(16);not null"`
}
