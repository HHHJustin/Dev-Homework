[toc]

# Model
## User table 
- ID: User table的primary key
- Username: 使用者登入帳號
- Password: 登入密碼
- Role: 區分User及Manager
- Todos: 建立與Todos的Foreign Key
```go
type User struct {
	ID       int    `gorm:"primaryKey"`
	Username string `gorm:"type:varchar(100);not null"`
	Password string `gorm:"type:varchar(100);not null"`
	Role     string `gorm:"type:varchar(100);not null"`
	Todos    []Todo `gorm:"foreignKey:UserID"`
}
```
```sql
+----------+--------------+------+-----+---------+----------------+
| Field    | Type         | Null | Key | Default | Extra          |
+----------+--------------+------+-----+---------+----------------+
| id       | bigint       | NO   | PRI | NULL    | auto_increment |
| username | varchar(100) | NO   |     | NULL    |                |
| password | varchar(100) | NO   |     | NULL    |                |
| role     | varchar(100) | NO   |     | NULL    |                |
+----------+--------------+------+-----+---------+----------------+
```

## Todo table
- ID: primaryKey
- Task: 輸入的任務
- Done: 是否完成任務
- UserID: 以此區分使用者的Todo list
```go
type Todo struct {
	ID     int    `gorm:"primaryKey"`
	Task   string `gorm:"type:varchar(100);not null"`
	Done   bool   `gorm:"default:false"`
	UserID uint   `gorm:"not null"`
	User   User   `gorm:"foreignKey:UserID;references:ID"`
}
```

```sql
+---------+--------------+------+-----+---------+----------------+
| Field   | Type         | Null | Key | Default | Extra          |
+---------+--------------+------+-----+---------+----------------+
| id      | bigint       | NO   | PRI | NULL    | auto_increment |
| task    | varchar(100) | NO   |     | NULL    |                |
| done    | tinyint(1)   | YES  |     | 0       |                |
| user_id | bigint       | NO   | MUL | NULL    |                |
+---------+--------------+------+-----+---------+----------------+
```

# Page
## Login
![image](https://hackmd.io/_uploads/B1TZBEusA.png)
- PostForm:
    - Username: 使用者帳號
    - Password: 使用者密碼
- Button:
    - Login: 登入，認證username及password後切換到"/todos"。
    - Register: 註冊，切換至"/register"。

## Register
![image](https://hackmd.io/_uploads/HJiXSN_jR.png)
- PostForm:
    - Username: 使用者帳號
    - Password: 使用者密碼
- Button:
    - Register: 認證username及password，如果username被使用過，會認證錯誤。

## Toto List 
![image](https://hackmd.io/_uploads/rJM3VEdjA.png)
- Table:
    - Logout: 登出
    - ID: Index。
    - Task: 寫入的內容。
    - Done: 用checkbox表示是否完成。
    - Delete: 按下表示刪除。
    - New: 在Enter new task中輸入新文字，按下Enter可以新增一筆資料。


# Router
/main.go
```go=
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
```
## GET - /
/handler/user.go
```go
// @Summary Render registration page
// @Description Displays the registration HTML page for new user registration
// @Tags pages
// @Accept html
// @Produce html
// @Success 200 {string} string "HTML page with registration form"
// @Router /register [get]
func RegisterPageHandler(c *gin.Context) {
	c.HTML(http.StatusOK, "register.html", nil)
}
```

## POST - /
```go=
// LoginHandler godoc
// @Summary Log in a user
// @Description Authenticate a user and return a JWT token
// @Tags users
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param username formData string true "Username"
// @Param password formData string true "Password"
// @Success 200 {object} handler.SuccessResponse "Login successful with JWT token"
// @Failure 401 {object} handler.ErrorResponse "Unauthorized - Can't find the user or invalid username or password"
// @Failure 500 {object} handler.ErrorResponse "Internal Server Error - Error creating token"
// @Router / [post]
func LoginHandler(c *gin.Context, db *gorm.DB) {
	username := c.PostForm("username")
	password := c.PostForm("password")
        // 抓取PostForm username及password的資料 
	var user database.User
	if err := db.Where("username = ?", username).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Message: "Can't find the user"})
		return
	}
        // 如果在user的table中找到username，則將資料放到user中，否則將回傳錯誤。
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Message: "Invalid username or password"})
		return
	}
        // 依據設定獲取token
	tokenString, err := token.CreateToken(username, user.ID)
	if err != nil {
		c.String(http.StatusInternalServerError, "Error creating token")
		return
	}
        // 將此token放入cookie中的"token"當中
	c.SetCookie("token", tokenString, 3600, "/", "localhost", false, true)
	c.Redirect(http.StatusSeeOther, "/todos")
}
```

- CreateToken
```go=
// Function to create JWT tokens with claims
func CreateToken(username string, user_id int) (string, error) {
        // claims中放入username,userid及過期時間等等重要資訊
	claims := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":    username,                         // Subject (user identifier)
		"iss":    "todo-app",                       // Issuer
		"aud":    "user",                           // Audience (user role)
		"exp":    time.Now().Add(time.Hour).Unix(), // Expiration time
		"iat":    time.Now().Unix(),                // Issued at
		"userid": user_id,
	})
        // 依據上面設定獲取完整token
	tokenString, err := claims.SignedString(secretKey)
	if err != nil {
		return "", err
	}
	// Print information about the created token
	fmt.Printf("Token claims added: %+v\n", claims)
	return tokenString, nil
}
```


## GET - /register
/handler/user.go
```go=
// @Summary Render registration page
// @Description Displays the registration HTML page for new user registration
// @Tags pages
// @Accept html
// @Produce html
// @Success 200 {string} string "HTML page with registration form"
// @Router /register [get]
func RegisterPageHandler(c *gin.Context) {
	c.HTML(http.StatusOK, "register.html", nil)
}
```
## POST - /register
/handler/user.go
```go=
// @Summary Register a new user
// @Description Register a new user with a username and password
// @Tags users
// @Accept  application/x-www-form-urlencoded
// @Produce json
// @Param username formData string true "Username"
// @Param password formData string true "Password"
// @Success 200 {object} handler.SuccessResponse "Registration successful"
// @Failure 409 {object} ErrorResponse "User already exists or password error"
// @Router /register [post]
func RegisterHandler(c *gin.Context, db *gorm.DB) {
	username := c.PostForm("username")
	password := c.PostForm("password")
        // 抓取PostForm中的數值。
	var user database.User
	if err := db.Where("username = ?", username).First(&user).Error; err == nil {
		c.JSON(http.StatusConflict, ErrorResponse{Message: "User already exist"})
		return
	}
        // 在database.User中尋找輸入的username，如果找到就把資料放到創建的user中。
	hashedpassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusConflict, ErrorResponse{Message: "Password type error"})
		return
	}
        // 將database中已經過hash function處理的password進行解碼。
	user = database.User{
		Username: username,
		Password: string(hashedpassword),
		Role:     "user",
	}
	db.Create(&user)
        // 創建user 
	c.Redirect(http.StatusFound, "/")
}
```

## Group - / 
/main.go
```go=
    ...
	// Todo
	authRoutes := router.Group("/").Use(func(c *gin.Context) {
		handler.AuthMiddleware(c)
	})
    ...
```

/handler/middleware.go
```go=
func AuthMiddleware(c *gin.Context) {
	// Retrieve the token from the cookie
	tokenString, err := c.Cookie("token")
	if err != nil {
		fmt.Println("Token missing in cookie")
		c.Redirect(http.StatusSeeOther, "/")
		c.Abort()
		return
	}

	// Verify the token
	token, err := token.VerifyToken(tokenString)
	if err != nil {
		fmt.Printf("Token verification failed: %v\\n", err)
		c.Redirect(http.StatusSeeOther, "/")
		c.Abort()
		return
	}

	// Print information about the verified token
	fmt.Printf("Token verified successfully. Claims: %+v\\n", token.Claims)

	// Continue with the next middleware or route handler
	c.Next()
}
```
- VerifyToken
```go=
func VerifyToken(tokenString string) (*jwt.Token, error) {
	// 檢查是否valid，合格就回傳，否則回傳失敗。
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})

	// Check for verification errors
	if err != nil {
		return nil, err
	}

	// Check if the token is valid
	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	// Return the verified token
	return token, nil
}
```

## GET - /todos (authrouter)
/handler/todolist.go
```go=
// @Summary Getting todo list
// @Description Getting all todos from todos table.
// @Tags todos
// @Accept  json
// @Produce  json
// @Success 200 {object} []database.TodoWithIndex
// @Failure 500 {object} ErrorResponse
// @Router /todos [get]
func GetTodoListHandler(c *gin.Context, db *gorm.DB) {
    // 從cookie中的"token"抓取token資訊
	tokenString, err := c.Cookie("token")
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Message: "Authorization token not provided"})
		return
	}
	// 轉換token資訊得到claims中的userID
	userID, err := token.ParseTokenAndGetUserID(tokenString)
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Message: "Invalid token"})
		return
	}
	var todos []database.Todo
	var todoWithIndex []database.TodoWithIndex
    // 抓取符合UserID的資料，如果找到入todos中
	result := db.Where("user_id = ?", userID).Find(&todos)
	log.Println(result)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Message: "failed to fetch data"})
		return
	}
    // 將資料依序放到todoWithIndex中
	for i, todo := range todos {
		todoWithIndex = append(todoWithIndex, database.TodoWithIndex{
			Index: i + 1,
			Todo:  todo,
		})
	}
    // 傳入todoWithIndex
	c.HTML(http.StatusOK, "todolist.html", todoWithIndex)
}
```

## POST - /todos (authrouter)
/database/model.go
```go=
type TodoWithIndex struct {
	Todo
	Index int
}
```

/handler/todolist.go
```go=
/// @Summary Create new task
// @Description Create new task then return task.ID
// @Tags todos
// @Accept  x-www-form-urlencoded
// @Produce  json
// @Param   task  formData  string  true  "task"
// @Success 201 {object} database.Todo
// @Failure 500 {object} ErrorResponse
// @Router /todos [post]
func TodoListHandler(c *gin.Context, db *gorm.DB) {
	// 從cookie中的"token"抓取token資訊
	tokenString, err := c.Cookie("token")
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Message: "Authorization token not provided"})
		return
	}
	// 轉換token資訊得到claims中的userID
	userID, err := token.ParseTokenAndGetUserID(tokenString)
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Message: "Invalid token"})
		return
	}
    // 抓取task中的資料
	task := c.PostForm("task")
    // 將task及userID放到todo中，Done統一是false
	todo := &database.Todo{
		Task:   task,
		Done:   false,
		UserID: userID,
	}
    // 在table中建立資料
	createResult := db.Create(&todo)
	if createResult.Error != nil {
		log.Fatal("Failed to insert new todo:", createResult.Error)
	}

	log.Println("New todo inserted with ID:", todo.ID)
    // 切換回"/todos"
	c.Redirect(http.StatusSeeOther, "/todos")
}
```

## POST - /updateTask (authrouter)
/handler/todolist.go
```go=
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
        // 抓取todo table中的id及task值
	id := c.PostForm("id")
	task := c.PostForm("task")
	var todo database.Todo
        // 尋找此筆資料是否存在todo table中，有則放到todo中
	if err := db.Find(&todo, id).Error; err != nil {
		panic("Can't find the item")
	}
        // 更新todo中的task
	todo.Task = task
        // 用db.Save存到todo中
	db.Save(todo)
	c.Redirect(http.StatusSeeOther, "/todos")
}
```

## POST - /updateDone (authrouter)
/handler/todolist.go
```go
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
```

## POST - /deleteTask (authrouter)
/handler/todolist.go
```go
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
```

## POST - /logout
/handler/user.go
```go=
func LogoutHandler(c *gin.Context) {
    // 將cookie中的token刪除
	c.SetCookie("token", "", -1, "/", "localhost", false, true)
	c.Redirect(http.StatusFound, "/")
}
```