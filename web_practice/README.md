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

# 加入資料庫 Ver.
## 創建資料庫
<!-- // 1. 創建Database - TodoList
// 2. 創建Table - TodoTable
	// - Id(Primary key, Int, auto_increment), 
	// - Task(string, required), 
	// - Done(bool, default:fasle).
	// - CreateAt
	// - UpdateAt	
// 3. 創建Gorm Model. -->
## CRUD
### R(Read)
<!-- 使用Get fetch web page時，直接從資料庫讀取資料從Table中第一筆依序列出在Page上。(Done以box方式顯示)
1. 建立可以顯示的HTML
2. 利用GET抓取HTML -->
### Create
<!-- 頁面表格的最下面，新增一個PostForm，當輸入文字按下Enter後，將Id, Task, Done輸入資料庫當中，輸入後更新頁面將資料從新顯示在畫面。 -->
### Update
<!-- 每一行旁設計更新按鈕，按下後跳出更改Task畫面。
-> 更新Task後更新頁面。
按下Done可直接更改是否完成狀態。 -->
### Delete
每一行旁邊設置Delete按鈕，按下之後直接刪除該筆資料。後更新頁面。