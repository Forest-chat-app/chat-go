package model

type RoomMembers struct {
	ID         string `gorm:"primaryKey;type:varchar(255)"`
	UserID     string `gorm:"type:varchar(255);not null"`
	RoomID     string `gorm:"type:varchar(255);not null"`
	JoinedAt   int64  `gorm:"not null"`
	LastReadId string `gorm:"type:varchar(255);"`
	UserRole   string `gorm:"type:varchar(255);not null"`
}
