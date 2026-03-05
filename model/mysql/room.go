package mysql

type Room struct {
	ID           string `gorm:"primaryKey;type:varchar(255)"`
	RoomName     string `gorm:"type:varchar(255);not null"`
	CreatorID    string `gorm:"foreignKey;type:varchar(255);not null"`
	Introduction string `gorm:"type:varchar(255);not null"`
	Tag          string `gorm:"type:varchar(255);not null"`
	Status       int16  `gorm:"type:int(16);not null"`
	CreatedAt    int64  `gorm:"not null"`
	UpdatedAt    int64  `gorm:"not null"`
	Avatar       string `gorm:"type:varchar(255)"`
	// Foreign key, GORM will handle this if you have the User model
	// Creator User `gorm:"foreignKey:CreatorID"`
}
