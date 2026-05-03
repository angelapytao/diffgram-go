package entity

type Base struct {
	ID int `gorm:"column:id;primarykey;autoIncrement"`
}
