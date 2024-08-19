# Todo List
[toc]
## 構想
Todo List function包含:
- todo中的column → id, task, done
- 顯示database中todos項目。
- 新增Task。
- 刪除Task。
- 更新表格中的Task。
- Done可以打勾以及取消。

1. 創建todos table + gorm model 

## Model
- Id: Primary Key
- Task: 100字以內的文字內容，不可為NULL。
- Done: 紀錄Task是否已經完成。
```go
type Todo struct {
	Id   int    `gorm:"primaryKey"`
	Task string `gorm:"type:varchar(100);not null"`
	Done bool   `gorm:"default:false"`
}
```
- MySQL中的todos table。
```sql
+-------+--------------+------+-----+---------+----------------+
| Field | Type         | Null | Key | Default | Extra          |
+-------+--------------+------+-----+---------+----------------+
| id    | bigint       | NO   | PRI | NULL    | auto_increment |
| task  | varchar(100) | NO   |     | NULL    |                |
| done  | tinyint(1)   | YES  |     | 0       |                |
+-------+--------------+------+-----+---------+----------------+
```


## HTML
```html
<body>
    <table>
        <caption>Todo List</caption>
        <thead>
            <tr>
                <th>ID</th>
                <th>Task</th>
                <th>Done</th>
            </tr>
        </thead>
        <tbody>
            {{range .}}
            <tr>
                <td>{{.Index}}</td>
                <td onclick="makeEditable(this, {{.Id}})">{{.Task}}</td>
                <td>
                    <input type="checkbox"
                           {{if .Done}}checked="checked"{{end}}
                           onchange="updateDoneStatus({{.Id}}, this)">
                </td>
                <td>
                    <form action="/deleteTask" method="post">
                        <input type="hidden" name="id" value="{{.Id}}">
                        <button type="submit">Delete</button>
                    </form>
                </td>
                
            </tr>
            {{else}}
            <tr>
                <td colspan="3">No tasks found</td>
            </tr>
            {{end}}
            <tr>
                <form action="/todos" method="post">
                    <td>New</td> 
                    <td><input type="text" name="task" placeholder="Enter new task" required></td>
                    <td><button type="submit">Add Task</button></td>
                </form>
            </tr>
        </tbody>
    </table>
</body>
```

![image](https://github.com/HHHJustin/Dev-Homework/blob/main/coffee_problem/images/image.jpg)
- 每筆資料包含:
    - ID: Index。
    - Task: 寫入的內容。
    - Done: 用checkbox表示是否完成。
    - Delete: 按下表示刪除。
- New: 在Enter new task中輸入新文字，按下Enter可以新增一筆資料。

## Router
### Main
```go=
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

```

### GET & POST - /todos

```go=

// database/model.go
type TodoWithIndex struct {
	Todo
	Index int
}

// handler/handler.go
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
```
- 為了讓todo list中的ID可以依照1, 2, 3...的順序打出來，新增一個TodoWithIndex，其中的Index在每次用GET開啟HTML網頁前，都會按照任務的創建順序重新從1開始排序。
- 如果Router Method是GET，則會先更新TodoWithIndex，再將其與todolist.html一起開啟。
- 如果Router Method是POST，代表要新增一項任務，這時會先抓取POST中的Value，並且創建一個Todo的架構，將此Value輸入到Todo架構中，最後使用gorm中的.Create在todos table中創建新的資料。用.Find來檢查是否有創建成功。最後用http.Redirect導回GET-./todos 中。

### POST - /updateTask
更新Done
```go
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
```
- 如果使用者點選todo中已存在的資料，並且輸入新的資料後，Serve端會抓取修改的id以及新的task，並且利用id找到在todos中的位置，利用db.Save修改task內容。

### POST - /updateDone
更新Task
```go
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
```
如果使用者點選todo list中已存在的資料後面的checkbox，Serve端會抓取修改的id，並且利用id找到在todos中的位置，利用db.Save修改done的狀態。

### POST - /deleteTask
刪除Task
```go
func DeleteTask(c *goo.Context, db *gorm.DB) {
	id := c.Req.FormValue("id")
	var todo database.Todo
	if err := db.Delete(&todo, id).Error; err != nil {
		panic("Can't find the item")
	}
	http.Redirect(c.Writer, c.Req, "/todos", http.StatusSeeOther)
}
```
如果使用者點選todo list中已存在的資料後面的Delete，Serve端會抓取修改的id，並且利用id找到在todos中的位置，利用db.Delete刪除。
