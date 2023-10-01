package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/library-management-system/library/controllers"
	"github.com/library-management-system/library/middlewares"
)

func HandleUserRoutes(incomingRoute *gin.Engine) {
	incomingRoute.Use(middlewares.AuthUser)
	incomingRoute.PATCH("/user/add-book", controllers.AddBook)
	incomingRoute.POST("/user/buy-book", controllers.BuyBook)
	incomingRoute.GET("/user/refresh-token", controllers.RefreshToken)
}
