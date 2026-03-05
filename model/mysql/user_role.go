package mysql

type UserRole struct {
	ID        string `gorm:"primaryKey;type:varchar(255)"`
	UserID    string `gorm:"type:varchar(255);not null"`
	RoleID    string `gorm:"type:varchar(255);not null"`
	CreatedAt int64  `gorm:"not null"`
	UpdatedAt int64  `gorm:"not null"`
}
