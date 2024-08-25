package handler

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"
	"todolist/database"

	"todolist/token"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func FormatAsDate(t time.Time) string {
	year, month, day := t.Date()
	return fmt.Sprintf("%d-%02d-%02d", year, month, day)
}

// @Summary Getting todo list
// @Description Getting all todos from todos table.
// @Tags todos
// @Accept  json
// @Produce  json
// @Success 200 {object} []database.TodoWithIndex
// @Failure 500 {object} ErrorResponse
// @Router /todos [get]
func GetTodoListHandler(c *gin.Context, db *gorm.DB) {
	tokenString, err := c.Cookie("token")
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Message: "Authorization token not provided"})
		return
	}
	// Get UserID from token
	userID, err := token.ParseTokenAndGetUserID(tokenString)
	log.Println("UserID = ", userID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Message: "Invalid token"})
		return
	}
	var todos []database.Todo
	var todoWithIndex []database.TodoWithIndex
	result := db.Where("user_id = ?", userID).Find(&todos)
	log.Println(result)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Message: "failed to fetch data"})
		return
	}
	for i, todo := range todos {
		todoWithIndex = append(todoWithIndex, database.TodoWithIndex{
			Index: i + 1,
			Todo:  todo,
		})
	}
	c.HTML(http.StatusOK, "todolist.html", todoWithIndex)
}

// @Summary Create new task
// @Description Create new task then return task.ID
// @Tags todos
// @Accept  x-www-form-urlencoded
// @Produce  json
// @Param   task  formData  string  true  "task"
// @Success 201 {object} database.Todo
// @Failure 500 {object} ErrorResponse
// @Router /todos [post]
func TodoListHandler(c *gin.Context, db *gorm.DB) {
	tokenString, err := c.Cookie("token")
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Message: "Authorization token not provided"})
		return
	}
	// Get UserID from token
	userID, err := token.ParseTokenAndGetUserID(tokenString)

	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Message: "Invalid token"})
		return
	}
	task := c.PostForm("task")
	todo := &database.Todo{
		Task:   task,
		Done:   false,
		UserID: userID,
	}
	createResult := db.Create(&todo)
	if createResult.Error != nil {
		log.Fatal("Failed to insert new todo:", createResult.Error)
	}

	log.Println("New todo inserted with ID:", todo.ID)
	c.Redirect(http.StatusSeeOther, "/todos")
}

// @Summary Update task
// @Description Update the task
// @Tags todos
// @Accept  x-www-form-urlencoded
// @Produce  json
// @Param   id    formData  string  true  "id"
// @Param   task  formData  string  true  "task"
// @Success 303 {string} string "Redirect to /todos"
// @Failure 404 {object} ErrorResponse "Can't find the task"
// @Failure 500 {object} ErrorResponse "Server error."
// @Router /updateTask [post]
func UpdateTask(c *gin.Context, db *gorm.DB) {
	id := c.PostForm("id")
	task := c.PostForm("task")
	var todo database.Todo
	if err := db.Find(&todo, id).Error; err != nil {
		panic("Can't find the item")
	}
	todo.Task = task
	db.Save(todo)
	c.Redirect(http.StatusSeeOther, "/todos")
}

// @Summary 更新任務的完成狀態
// @Description 根據任務 ID 更新任務的完成狀態 (done)
// @Tags todos
// @Accept  x-www-form-urlencoded
// @Produce  json
// @Param   id    formData  string  true  "任務的 ID"
// @Param   done  formData  bool    true  "完成狀態 (true/false)"
// @Success 200 {string} string "成功更新任務狀態"
// @Failure 400 {object} ErrorResponse "無效的完成狀態值"
// @Failure 404 {object} ErrorResponse "找不到任務"
// @Failure 500 {object} ErrorResponse "伺服器錯誤"
// @Router /updateDone [post]
func UpdateDoneCheckbox(c *gin.Context, db *gorm.DB) {
	id := c.PostForm("id")
	doneValue := c.PostForm("done")
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

// DeleteTask godoc
// @Summary 刪除任務
// @Description 根據任務 ID 刪除任務
// @Tags todos
// @Accept  x-www-form-urlencoded
// @Produce  json
// @Param   id  formData  string  true  "任務的 ID"
// @Success 303 {string} string "重定向到 /todos"
// @Failure 404 {object} ErrorResponse "找不到任務"
// @Failure 500 {object} ErrorResponse "伺服器錯誤"
// @Router /deleteTask [post]
func DeleteTask(c *gin.Context, db *gorm.DB) {
	id := c.PostForm("id")
	var todo database.Todo
	if err := db.Delete(&todo, id).Error; err != nil {
		panic("Can't find the item")
	}
	c.Redirect(http.StatusSeeOther, "/todos")
}
