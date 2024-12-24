package db

import (
	"database/sql"
	"fmt"
)

type ConfigDB struct {
	User     string
	Password string
	Name     string
	Host     string
	Port     string
}

var Conn *sql.DB

func Init(conf *ConfigDB) error {
	connStr := fmt.Sprintf(
		"user=%s password=%s dbname=%s host=%s port=%s sslmode=disable",
		conf.User,
		conf.Password,
		conf.Name,
		conf.Host,
		conf.Port,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("error connecting to the database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return fmt.Errorf("error pinging the database: %w", err)
	}

	Conn = db
	fmt.Println("successfully connected to the PostgreSQL database!")
	return nil
}
