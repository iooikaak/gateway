package mysql

// Token token
type Token struct {
	ID            int64  `gorm:"column:id;primary_key"`
	UserID        int64  `gorm:"column:user_id"`
	Token         string `gorm:"column:token"`
	TokenGenTime  int64  `gorm:"column:tokenGenTime"`
	Status        int64  `gorm:"column:status"`
	LastApiUpdate int64  `gorm:"column:last_api_update"`
}

// TableName table name
func (token *Token) TableName() string {
	return "app_user_token"
}
