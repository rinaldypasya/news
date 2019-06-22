package config

import (
	"fmt"
	"os"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"

	"github.com/rinaldypasya/news/news"
)

var (
	dbHost = os.Getenv("POSTGRES_HOST")
	dbUser = os.Getenv("POSTGRES_USER")
	dbPass = os.Getenv("POSTGRES_PASSWORD")
	dbName = os.Getenv("POSTGRES_DB")
	dbPort = "5432"
)

// DBInit create connection to database
func DBInit() *gorm.DB {
	connection := fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s", dbHost, dbPort, dbUser, dbPass, dbName)
	db, err := gorm.Open("postgres", connection)
	if err != nil {
		panic("failed to connect to database")
	}

	db.AutoMigrate(&news.News{})
	return db
}
