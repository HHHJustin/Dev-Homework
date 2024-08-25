package main

import (
	"text/template"
	"todolist/database"
	_ "todolist/docs"
	"todolist/handler"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title Todo API
// @version 1.0
// @description This is a todo list API
// @host localhost:9999
// @BasePath /
func main() {
	db, err := database.ConnectDatabase()
	db.AutoMigrate(&database.Todo{})
	if err != nil {
		panic("Connect Fail!!")
	}

	router := gin.New()
	router.Use(gin.Logger())
	router.SetFuncMap(template.FuncMap{
		"FormatAsDate": handler.FormatAsDate,
	})
	router.LoadHTMLGlob("templates/*")
	router.Static("/assets", "./static")

	// Login
	router.GET("/", func(c *gin.Context) {
		handler.LoginPageHandler(c)
	})
	router.POST("/", func(c *gin.Context) {
		handler.LoginHandler(c, db)
	})

	// Register
	router.GET("/register", func(c *gin.Context) {
		handler.RegisterPageHandler(c)
	})
	router.POST("/register", func(c *gin.Context) {
		handler.RegisterHandler(c, db)
	})

	// Todo
	authRoutes := router.Group("/").Use(func(c *gin.Context) {
		handler.AuthMiddleware(c)
	})
	authRoutes.GET("/todos", func(c *gin.Context) {
		handler.GetTodoListHandler(c, db)
	})
	authRoutes.POST("/todos", func(c *gin.Context) {
		handler.TodoListHandler(c, db)
	})
	authRoutes.POST("/updateDone", func(c *gin.Context) {
		handler.UpdateDoneCheckbox(c, db)
	})
	authRoutes.POST("/updateTask", func(c *gin.Context) {
		handler.UpdateTask(c, db)
	})
	authRoutes.POST("/deleteTask", func(c *gin.Context) {
		handler.DeleteTask(c, db)
	})

	// Log out
	router.POST("/logout", handler.LogoutHandler)
	// Swagger documentation route
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	router.Run(":9999")
}
