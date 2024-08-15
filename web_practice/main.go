package main

import (
	"fmt"
	"goo"
	"html/template"
	"net/http"
	"time"
)

type student struct {
	Name string
	Age  int8
}

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

func main() {
	r := goo.New()
	r.Use(goo.Logger())
	r.SetFuncMap(template.FuncMap{
		"FormatAsDate": FormatAsDate,
	})
	r.LoadHTMLGlob("templates/*")
	r.Static("/assets", "./static")

	stu1 := &student{Name: "gootutu", Age: 20}
	stu2 := &student{Name: "Jack", Age: 22}
	r.GET("/", func(c *goo.Context) {
		c.HTML(http.StatusOK, "css.tmpl", nil)
	})
	r.GET("/students", func(c *goo.Context) {
		c.HTML(http.StatusOK, "arr.tmpl", goo.H{
			"title":  "goo",
			"stuArr": [2]*student{stu1, stu2},
		})
	})

	r.GET("/date", func(c *goo.Context) {
		c.HTML(http.StatusOK, "custom_func.tmpl", goo.H{
			"title": "goo",
			"now":   time.Date(2019, 8, 17, 0, 0, 0, 0, time.UTC),
		})
	})

	r.GET("/login", login)
	r.POST("/login", login)

	r.Run(":9999")
}
