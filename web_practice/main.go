package main

import (
	"fmt"
	"goo"
	"html/template"
	"net/http"
	"time"
)

func FormatAsDate(t time.Time) string {
	year, month, day := t.Date()
	return fmt.Sprintf("%d-%02d-%02d", year, month, day)
}

func login(c *goo.Context) {
	fmt.Println("method:", c.Method) //取得請求的方法
	if c.Method == "GET" {
		c.HTML(http.StatusOK, "login.gtpl", nil)
	} else {
		//請求的是登入資料，那麼執行登入的邏輯判斷
		fmt.Println("username:", c.PostForm("username"))
		fmt.Println("password:", c.PostForm("password"))
	}
}

// 需要編號及任務內容
type Todo struct {
	Id   int    // 編號
	Task string // 內容
}

var (
	todoList      = []Todo{}
	todoListIndex = 1
)

func todoListHandler(c *goo.Context) {
	// 當使用GET時，fetch todo.html來渲染
	if c.Method == "GET" {
		c.HTML(http.StatusOK, "todo.html", goo.H{
			"Todos": todoList,
		})
	} else if c.Method == "POST" {
		// 當使用當中的POST時，抓取空格內的字串當成task
		task := c.PostForm("task")
		// 放入Todo當中並且依據原有的todoListIndex編號
		newTodo := Todo{
			Id:   todoListIndex,
			Task: task,
		}
		todoListIndex++
		// 放到原有的todolist中，且與html檔一起回傳
		todoList = append(todoList, newTodo)

		c.HTML(http.StatusOK, "todo.html", goo.H{
			"Todos": todoList,
		})
	}
}

func main() {
	r := goo.New()
	r.Use(goo.Logger())
	r.SetFuncMap(template.FuncMap{
		"FormatAsDate": FormatAsDate,
	})
	r.LoadHTMLGlob("templates/*")
	r.Static("/assets", "./static")

	r.GET("/", func(c *goo.Context) {
		c.HTML(http.StatusOK, "css.tmpl", nil)
	})

	r.GET("/login", login)
	r.POST("/login", login)
	r.GET("/todo", todoListHandler)
	r.POST("/todo", todoListHandler)
	r.Run(":9999")
}
