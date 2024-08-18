package handler

import (
	"fmt"
	"goo"
	"log"
	"main/database"
	"net/http"
	"strconv"
	"time"

	"gorm.io/gorm"
)

func FormatAsDate(t time.Time) string {
	year, month, day := t.Date()
	return fmt.Sprintf("%d-%02d-%02d", year, month, day)
}

func TodoListHandler(c *goo.Context, db *gorm.DB) {
	// 從todos中抓取資料
	var todos []database.Todo
	result := db.Find(&todos)
	if result.Error != nil {
		panic("failed to fetch data")
	}
	if c.Method == "GET" {
		var todoWithIndex []database.TodoWithIndex
		for i, todo := range todos {
			todoWithIndex = append(todoWithIndex, database.TodoWithIndex{
				Index: i + 1,
				Todo:  todo,
			})
		}
		c.HTML(http.StatusOK, "todolist.html", todoWithIndex)
	} else {
		task := c.Req.FormValue("task")
		todo := &database.Todo{
			Task: task,
		}
		createResult := db.Create(&todo)
		if createResult.Error != nil {
			log.Fatal("Failed to insert new todo:", result.Error)
		}
		log.Println("New todo inserted with ID:", todo.Id)
		result = db.Find(&todos)
		if result.Error != nil {
			panic("failed to fetch data")
		}
		http.Redirect(c.Writer, c.Req, "/todos", http.StatusSeeOther)
	}
}

func UpdateTask(c *goo.Context, db *gorm.DB) {
	id := c.Req.FormValue("id")
	task := c.Req.FormValue("task")
	var todo database.Todo
	if err := db.Find(&todo, id).Error; err != nil {
		panic("Can't find the item")
	}
	todo.Task = task
	db.Save(todo)
	http.Redirect(c.Writer, c.Req, "/todos", http.StatusSeeOther)
}

func UpdateDoneCheckbox(c *goo.Context, db *gorm.DB) {
	id := c.Req.FormValue("id")
	doneValue := c.Req.FormValue("done")
	done, err := strconv.ParseBool(doneValue)
	if err != nil {
		panic("Invalid done value")
	}
	var todo database.Todo
	if err := db.Find(&todo, id).Error; err != nil {
		panic("Can't find the item")
	}
	todo.Done = done
	db.Save(todo)
}

func DeleteTask(c *goo.Context, db *gorm.DB) {
	id := c.Req.FormValue("id")
	var todo database.Todo
	if err := db.Delete(&todo, id).Error; err != nil {
		panic("Can't find the item")
	}
	http.Redirect(c.Writer, c.Req, "/todos", http.StatusSeeOther)
}
