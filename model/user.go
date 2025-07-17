package model

type User struct {
	ID          string `gorm:"primaryKey;type:varchar(255)"`
	UserAccount string `gorm:"unique;type:varchar(255);not null"`
	Password    string `gorm:"type:varchar(255);not null"`
	Nickname    string `gorm:"type:varchar(255);"`
	Email       string `gorm:"type:varchar(255);"`
	Avatar      string `gorm:"type:varchar(255);"`
	CreatedAt   int64  `gorm:"not null"`
	UpdatedAt   int64  `gorm:"not null"`
}
