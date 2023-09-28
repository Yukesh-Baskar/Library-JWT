package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/library-management-system/library/controllers"
)

func HandleUserAuthRoutes(incomingRoutes *gin.Engine) {
	incomingRoutes.POST("/user/sign-up", controllers.SignUp)
	incomingRoutes.GET("/user/login", controllers.Login)
	//
	// wef
}
