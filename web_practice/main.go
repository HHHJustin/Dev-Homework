package main

import (
	"goo"
	"main/database"
	"main/handler"
	"text/template"
)

func main() {
	db, err := database.ConnectDatabase()
	db.AutoMigrate(&database.Todo{})
	if err != nil {
		panic("Connect Fail!!")
	}
	router := goo.New()
	router.Use(goo.Logger())
	router.SetFuncMap(template.FuncMap{
		"FormatAsDate": handler.FormatAsDate,
	})
	router.LoadHTMLGlob("templates/*")
	router.Static("/assets", "./static")
	router.GET("/todos", func(c *goo.Context) {
		handler.TodoListHandler(c, db)
	})
	router.POST("/todos", func(c *goo.Context) {
		handler.TodoListHandler(c, db)
	})
	router.POST("/updateDone", func(c *goo.Context) {
		handler.UpdateDoneCheckbox(c, db)
	})
	router.POST("/updateTask", func(c *goo.Context) {
		handler.UpdateTask(c, db)
	})
	router.POST("/deleteTask", func(c *goo.Context) {
		handler.DeleteTask(c, db)
	})
	router.Run(":9999")
}
