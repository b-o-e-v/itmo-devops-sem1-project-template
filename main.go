package main

import (
	"log"
	"os"
	"project_sem/pkg/db"
	"project_sem/pkg/envs"
	"project_sem/server"

	_ "github.com/lib/pq"
)

func init() {
	// загружаем ENV
	if err := envs.Init(); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	// подключаемся к DB
	configDB := &db.ConfigDB{
		User:     envs.Config.UserDB,
		Password: envs.Config.PasswordDB,
		Name:     envs.Config.NameDB,
		Host:     envs.Config.HostDB,
		Port:     envs.Config.PortDB,
	}

	if err := db.Init(configDB); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}

func main() {
	// запускаем сервер
	if err := server.Up(envs.Config.Port); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	// закрываем соединение
	defer db.Conn.Close()
}
