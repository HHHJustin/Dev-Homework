# Todo List
## Todo 結構
```go
type Todo struct {
	Id   int    // 編號
	Task string // 內容
}
```
- 輸入的內容與其編號
## Handler
```go
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
        // 依據或許的TodoList將傳到HTML渲染
		c.HTML(http.StatusOK, "todo.html", goo.H{
			"Todos": todoList,
		})
	}
}

//...
func main(){
//...
    r.GET("/todo", todoListHandler) 
	r.POST("/todo", todoListHandler)
    r.RUN(":9999")
}

```
- todoList:存放目前為止所輸入的Todo。
- todoListIndex:紀錄目前為止有多少Todo。

## HTML內容
```html
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Todo List</title>
</head>
<body>
    <h1>Todo List</h1>
    <form action="/todo" method="post">
        <input type="text" name="task" placeholder="New task">
        <button type="submit">Add</button>
    </form>
    <ul>
        {{ range .Todos }}
        <li>{{ .Id }}. {{ .Task }}</li>
        {{ end }}
    </ul>
</body>
</html>
```

- 在</form>中使用{{ range .Todos }}方式，將所得到的TodoList依序顯示