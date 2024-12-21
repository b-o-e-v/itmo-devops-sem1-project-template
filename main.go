package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type App struct {
	port       string
	dbName     string
	dbUser     string
	dbPassword string
	dbHost     string
	dbPort     string
}

var app *App
var db *sql.DB

func init() {
	// Загружаем переменные окружения из файла .env
	err := godotenv.Load()
	if err != nil {
		log.Fatal("- error loading .env file")
		os.Exit(1)
	}

	app = &App{
		port:       os.Getenv("APP_PORT"),
		dbName:     os.Getenv("POSTGRES_DB"),
		dbUser:     os.Getenv("POSTGRES_USER"),
		dbPassword: os.Getenv("POSTGRES_PASSWORD"),
		dbHost:     os.Getenv("POSTGRES_HOST"),
		dbPort:     os.Getenv("POSTGRES_PORT"),
	}

	// Подключаемся к postgres
	connStr := fmt.Sprintf(
		"user=%s password=%s dbname=%s host=%s port=%s sslmode=disable",
		app.dbUser,
		app.dbPassword,
		app.dbName,
		app.dbHost,
		app.dbPort,
	)

	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("- error connecting to the database: ", err)
		os.Exit(1)
	}
	defer db.Close()

	// Проверяем подключение
	err = db.Ping()
	if err != nil {
		log.Fatal("- error pinging the database: ", err)
		os.Exit(1)
	}

	fmt.Println("successfully connected to the PostgreSQL database!")
}

func main() {
	r := gin.Default()

	r.GET("/ping", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{"data": "pong"})
	})

	r.GET("/api/v0/prices", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{"data": "data.csv"})
	})

	r.POST("/api/v0/prices", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{
			"total_items":      100,
			"total_categories": 15,
			"total_price":      100000,
		})
	})

	r.Run(fmt.Sprintf(":%s", app.port))
}
