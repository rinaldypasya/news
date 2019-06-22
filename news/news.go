package news

import (
	"time"
)

type News struct {
	ID      int
	Author  string `gorm:"type:text"`
	Body    string `gorm:"type:text"`
	Created time.Time
}
