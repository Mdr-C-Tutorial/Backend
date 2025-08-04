package main

import (
	"fmt"
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
	
	// 添加调试信息
	fmt.Printf("Loaded JWT Config: %+v\n", config.AppConfig.JWT)

	database.Init()

	r := gin.Default()

	// 添加 CORS 中间件
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"}, // 允许前端开发服务器的域名
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		AllowCredentials: true, // 允许携带 cookie
	}))

	// 添加搜索路由
	r.GET("/search/:query", handlers.HandleSearch)

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
		user := api.Group("/user")
		{
			user.PATCH("/name/:id", middleware.AuthRequired(), handlers.UpdateUsername(database.DB))
			user.DELETE("/delete/:id", middleware.AuthRequired(), handlers.HandleDeleteUser(database.DB)) // 添加删除用户路由
		}
		api.POST("/feedback", middleware.AuthRequired(), handlers.HandleFeedback)
	}
	if err := r.Run(":8080"); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
