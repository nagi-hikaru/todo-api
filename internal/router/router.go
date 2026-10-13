package router

import (
	"github.com/gin-gonic/gin"
	_ "github.com/nagi-hikaru/todo-api/docs"
	"github.com/nagi-hikaru/todo-api/internal/database"
	"github.com/nagi-hikaru/todo-api/internal/handler"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func SetupHandler(r *gin.Engine, dbService database.DatabaseService) {
	h := handler.NewTodoHandler(dbService)
	api := r.Group("/api/v1")
	{
		api.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

		todos := api.Group("/todos")
		{
			todos.GET("", h.ListTodosHandler)
			todos.POST("", h.CreateTodoHandler)
			todos.GET("/:id", h.GetTodoHandler)
			todos.PUT("/:id", h.UpdateTodoHandler)
			todos.DELETE("/:id", h.DeleteTodoHandler)
			todos.PATCH("/:id/status", h.UpdateStatusHandler)
		}
	}
}
