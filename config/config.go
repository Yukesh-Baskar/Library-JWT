package config

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/library-management-system/library/routes"
)

func ConfigApp() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("error occured while loading env: %s \n", err.Error())
	}

	router := gin.New()
	router.Use(gin.Logger())
	routes.HandleUserAuthRoutes(router)
	routes.HandleUserRoutes(router)
	log.Fatal(router.Run(os.Getenv("PORT")))
}
