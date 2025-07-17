package model

type Room struct {
	ID        string `gorm:"primaryKey;type:varchar(255)"`
	RoomName  string `gorm:"type:varchar(255);not null"`
	CreatorID string `gorm:"type:varchar(255);not null"`
	IsPrivate bool   `gorm:"type:tinyint(1);not null"` // 使用 bool 映射 tinyint(1)
	IsDelete  bool   `gorm:"type:tinyint(1);not null"`
	CreatedAt int64  `gorm:"not null"`
	UpdatedAt int64  `gorm:"not null"`
	// Foreign key, GORM will handle this if you have the User model
	Creator User `gorm:"foreignKey:CreatorID"`
}
