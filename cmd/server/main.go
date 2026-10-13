package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/nagi-hikaru/todo-api/internal/database"
	"github.com/nagi-hikaru/todo-api/internal/middleware"
	"github.com/nagi-hikaru/todo-api/internal/router"
)

// @title Todo API
// @version 1.0
// @description 基于 Gin 和 SQLite 的待办事项 API
// @BasePath /api/v1
func main() {
	db, err := database.InitDatabase("./todos.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	dbService := database.NewDatabaseService(db)

	r := gin.New()
	middleware.SetupMiddleware(r)
	router.SetupHandler(r, dbService)

	if err := r.Run(":8080"); err != nil {
		log.Printf("启动 HTTP 服务失败: %v", err)
		return
	}
}
