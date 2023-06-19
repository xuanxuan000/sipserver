package gorm

// Model base model definition, including fields `ID`, `CreatedAt`, `UpdatedAt`, `DeletedAt`, which could be embedded in your models
//    type User struct {
//      gorm.Model
//    }
type Model struct {
	ID        uint   `json:"id" gorm:"primary_key"`
	CreatedAt int64  `json:"addtime" gorm:"column:addtime"`
	UpdatedAt int64  `json:"uptime" gorm:"column:uptime"`
	DeletedAt *int64 `json:"deltime" sql:"index" gorm:"column:deltime"`
}
