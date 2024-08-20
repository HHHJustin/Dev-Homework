# Todo List
# Model
```go
type Todo struct {
	Id   int    `gorm:"primaryKey"`
	Task string `gorm:"type:varchar(100);not null"`
	Done bool   `gorm:"default:false"`
}
```
```sql
+-------+--------------+------+-----+---------+----------------+
| Field | Type         | Null | Key | Default | Extra          |
+-------+--------------+------+-----+---------+----------------+
| id    | bigint       | NO   | PRI | NULL    | auto_increment |
| task  | varchar(100) | NO   |     | NULL    |                |
| done  | tinyint(1)   | YES  |     | 0       |                |
+-------+--------------+------+-----+---------+----------------+
```
- Id: 
- Task:
- Done:
# HTML
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

# Router
## GET & POST - /todos
顯示網頁 + 新增Task
```go
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

## POST - /updateTask
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
如果使用者點選todo list中已存在的資料後面的checkbox，Serve端會抓取修改的id，並且利用id找到在todos中的位置，利用db.Save修改done的狀態。

## POST - /updateDone
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

## POST - /deleteTask
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
