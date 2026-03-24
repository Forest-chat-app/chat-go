package mysql

type User struct {
	ID          string `gorm:"primaryKey;type:varchar(255)" json:"id"`
	UserAccount string `gorm:"unique;type:varchar(255);not null" json:"user_account"`
	Password    string `gorm:"type:varchar(255);not null" json:"-"`
	Nickname    string `gorm:"type:varchar(255)" json:"nickname"`
	Email       string `gorm:"type:varchar(255)" json:"email"`
	Avatar      string `gorm:"type:varchar(255)" json:"avatar"`
	Status      int16  `gorm:"type:smallint;not null;default:1" json:"status"`
	CreatedAt   int64  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   int64  `gorm:"autoUpdateTime" json:"updated_at"`
}
