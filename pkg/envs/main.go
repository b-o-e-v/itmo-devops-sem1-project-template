package envs

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Envs struct {
	Port       string
	UserDB     string
	PasswordDB string
	NameDB     string
	HostDB     string
	PortDB     string
}

var Config = &Envs{}

func Init() error {
	if err := godotenv.Load(); err != nil {
		return fmt.Errorf("error loading .env file: %w", err)
	}

	Config = &Envs{
		Port:       os.Getenv("APP_PORT"),
		UserDB:     os.Getenv("POSTGRES_USER"),
		PasswordDB: os.Getenv("POSTGRES_PASSWORD"),
		NameDB:     os.Getenv("POSTGRES_DB"),
		HostDB:     os.Getenv("POSTGRES_HOST"),
		PortDB:     os.Getenv("POSTGRES_PORT"),
	}

	fmt.Println("successfully loaded environment variables!")
	return nil
}
