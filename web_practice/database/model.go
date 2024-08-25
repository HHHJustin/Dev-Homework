package database

type User struct {
	ID       int    `gorm:"primaryKey"`
	Username string `gorm:"type:varchar(100);not null"`
	Password string `gorm:"type:varchar(100);not null"`
	Role     string `gorm:"type:varchar(100);not null"`
	Todos    []Todo `gorm:"foreignKey:UserID"`
}

type Todo struct {
	ID     int    `gorm:"primaryKey"`
	Task   string `gorm:"type:varchar(100);not null"`
	Done   bool   `gorm:"default:false"`
	UserID uint   `gorm:"not null"`
	User   User   `gorm:"foreignKey:UserID;references:ID"`
}

type TodoWithIndex struct {
	Todo
	Index int
}
