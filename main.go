package main

import (
	"log"

	"mdr/config"
	"mdr/database"
	"mdr/handlers"
	"mdr/middleware"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	config.Init()
	database.Init()

	r := gin.Default()

	// 添加 CORS 中间件
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"}, // 允许前端开发服务器的域名
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		AllowCredentials: true, // 允许携带 cookie
	}))

	api := r.Group("/api")
	{
		auth := api.Group("/auth")
		{
			auth.GET("/login", handlers.HandleCheckLogin)
			auth.POST("/login", handlers.HandleLogin)
			auth.POST("/register", handlers.HandleRegister)
			auth.GET("/verify-email", handlers.HandleVerifyEmail)
			auth.DELETE("/logout", middleware.AuthRequired(), handlers.HandleLogout)
		}
	}

	if err := r.Run(":8080"); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
