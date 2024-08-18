package database

// 需要編號及任務內容
type Todo struct {
	Id   int    `gorm:"primaryKey"`
	Task string `gorm:"type:varchar(100);not null"`
	Done bool   `gorm:"default:false"`
}

type TodoWithIndex struct {
	Todo
	Index int
}
