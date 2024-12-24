package main

import (
	"log"
	"os"
	"project_sem/pkg/db"
	"project_sem/pkg/env"
	"project_sem/server"

	_ "github.com/lib/pq"
)

func init() {
	// загружаем ENV
	if err := env.Init(); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	// подключаемся к DB
	configDB := &db.ConfigDB{
		User:     env.Config.UserDB,
		Password: env.Config.PasswordDB,
		Name:     env.Config.NameDB,
		Host:     env.Config.HostDB,
		Port:     env.Config.PortDB,
	}

	if err := db.Init(configDB); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}

func main() {
	// запускаем сервер
	if err := server.Up(env.Config.Port); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	// закрываем соединение
	defer db.Conn.Close()
}
