package server

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

func Up(port string) error {
	r := gin.Default()

	// получить данные из базы
	r.GET("/api/v0/prices", uploadArchive)
	// загрузить архив в базу
	r.POST("/api/v0/prices", loadArchive)

	if err := r.Run(fmt.Sprintf(":%s", port)); err != nil {
		return fmt.Errorf("failed to Listen and Serve: %w", err)
	}

	return nil
}
