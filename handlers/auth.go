package handlers

import (
	"fmt"
	"net/http"

	"mdr/database"
	"mdr/models"
	"mdr/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Remember bool   `json:"remember"`
}

type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

func HandleLogin(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求参数"})
		return
	}

	var user models.User
	result := database.DB.Where("username = ? OR email = ?", req.Username, req.Username).First(&user)
	if result.Error != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户名或密码错误"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户名或密码错误"})
		return
	}

	// 生成 token
	token, err := utils.GenerateToken(user.ID.String(), req.Remember)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "生成token失败"})
		return
	}

	// 设置 cookie
	maxAge := utils.GetCookieMaxAge(req.Remember)
	c.SetCookie("token", token, maxAge, "/", "", false, true) // 开发环境暂时设置 Secure 为 false

	c.JSON(http.StatusOK, gin.H{
		"user": gin.H{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
			"verified": user.Verified,
		},
	})
}

func HandleCheckLogin(c *gin.Context) {
	token, err := c.Cookie("token")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	userID, err := utils.ValidateToken(token)
	fmt.Println(err)
	if err != nil {
		c.SetCookie("token", "", -1, "/", "", false, true) // 清除无效的 token
		c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的登录状态"})
		return
	}

	var user models.User
	if err := database.DB.First(&user, "id = ?", userID).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": gin.H{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
			"role":     user.Role,
			"verified": user.Verified,
		},
	})
}

func HandleRegister(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求参数"})
		return
	}

	// 检查用户名和邮箱是否已存在
	var existingUser models.User
	if database.DB.Where("email = ?", req.Email).First(&existingUser).Error == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "邮箱已被注册"})
		return
	}

	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "密码加密失败"})
		return
	}
	user := models.User{
		Username: req.Username,
		Email:    req.Email,
		Password: string(hashedPassword),
		Role:     "user", // 添加默认角色
	}
	if err := database.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建用户失败"})
		return
	}

	// 生成验证邮件
	if err := utils.SendVerificationEmail(user); err != nil {
		c.JSON(http.StatusOK, gin.H{"message": "用户创建成功，但发送验证邮件失败"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "注册成功，请查收验证邮件"})
}

func HandleVerifyEmail(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少验证token"})
		return
	}

	userID, err := utils.VerifyEmailToken(token)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的验证链接"})
		return
	}

	uid, err := uuid.Parse(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}

	result := database.DB.Model(&models.User{}).Where("id = ?", uid).Update("verified", true)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "验证失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "邮箱验证成功"})
}

func HandleLogout(c *gin.Context) {
	c.SetCookie("token", "", -1, "/", "", true, true)
	c.JSON(http.StatusOK, gin.H{"message": "登出成功"})
}
