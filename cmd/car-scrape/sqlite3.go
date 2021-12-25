package main

import (
	"fmt"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Database struct {
	DB *gorm.DB
}

func Connect() (*Database, error) {
	db, err := gorm.Open(sqlite.Open("cars.db"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	return &Database{DB: db}, nil
}

func (d *Database) Close() {
	sql, err := d.DB.DB()
	if err != nil {
		fmt.Println("error while getting db")
		return
	}
	err = sql.Close()
	if err != nil {
		fmt.Println("error while closing sql db")
		return
	}
}
