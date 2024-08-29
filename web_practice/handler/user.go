package handler

import (
	"net/http"
	"todolist/database"

	"todolist/token"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

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
	var user database.User
	if err := db.Where("username = ?", username).First(&user).Error; err == nil {
		c.JSON(http.StatusConflict, ErrorResponse{Message: "User already exist"})
		return
	}
	hashedpassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusConflict, ErrorResponse{Message: "Password type error"})
		return
	}
	user = database.User{
		Username: username,
		Password: string(hashedpassword),
		Role:     "user",
	}
	db.Create(&user)
	c.Redirect(http.StatusFound, "/")
}

// @Summary Render login page
// @Description Displays the login HTML page for user authentication
// @Tags pages
// @Accept html
// @Produce html
// @Success 200 {string} string "HTML page with login form"
// @Router /login [get]
func LoginPageHandler(c *gin.Context) {
	c.HTML(http.StatusOK, "login.html", nil)
}

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
	var user database.User
	if err := db.Where("username = ?", username).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Message: "Can't find the user"})
		return
	}
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Message: "Invalid username or password"})
		return
	}
	tokenString, err := token.CreateToken(username, user.ID)
	if err != nil {
		c.String(http.StatusInternalServerError, "Error creating token")
		return
	}
	c.SetCookie("token", tokenString, 3600, "/", "127.0.0.1", false, true)
	c.Redirect(http.StatusSeeOther, "/todos")
}

func LogoutHandler(c *gin.Context) {
	c.SetCookie("token", "", -1, "/", "127.0.0.1", false, true)
	c.Redirect(http.StatusFound, "/")
}
