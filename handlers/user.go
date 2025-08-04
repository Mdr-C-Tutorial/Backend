package handlers

import (
	"net/http"

	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"mdr/database"
	"mdr/models"
)

type UpdateUsernameRequest struct {
	Username string `json:"username" binding:"required,min=3,max=20"`
}

type DeleteUserRequest struct {
	Confirm bool `json:"confirm" binding:"required"`
}

// UpdateUsername 允许管理员或用户本人通过提供用户 ID 来更新用户名
func UpdateUsername(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取目标用户ID
		targetUserID, err := uuid.Parse(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid target user ID"})
			return
		}

		// 从中间件获取当前用户ID (interface{})
		currentUserIDInterface, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
			return
		}

		// 打印获取到的 currentUserIDInterface 的信息
		fmt.Printf("DEBUG: currentUserIDInterface type: %T, value: %v\n", currentUserIDInterface, currentUserIDInterface)

		// 将当前用户ID interface{} 转换为 string 类型
		currentUserIDStr, ok := currentUserIDInterface.(string)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to get current user ID string"})
			return
		}
		fmt.Printf("DEBUG: currentUserIDStr: %s\n", currentUserIDStr)

		// 将当前用户ID字符串转换为 uuid.UUID 类型
		parsedCurrentUserID, err := uuid.Parse(currentUserIDStr)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to parse current user ID"})
			return
		}
		fmt.Printf("DEBUG: parsedCurrentUserID: %s\n", parsedCurrentUserID.String())

		// 查询当前用户信息
		var currentUser models.User
		if err := db.First(&currentUser, "id = ?", parsedCurrentUserID).Error; err != nil {
			fmt.Printf("DEBUG: Failed to find current user with ID %s: %v\n", parsedCurrentUserID.String(), err)
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Current user not found"})
			return
		}
		fmt.Printf("DEBUG: Found current user: ID=%s, Username=%s, Role=%s\n", currentUser.ID.String(), currentUser.Username, currentUser.Role)

		// 打印目标用户ID
		fmt.Printf("DEBUG: targetUserID: %s\n", targetUserID.String())

		// 验证权限: 当前用户是目标用户本人或管理员
		hasPermission := currentUser.ID == targetUserID || currentUser.Role == "admin"
		fmt.Printf("DEBUG: Permission check - Current User ID: %s, Target User ID: %s, Role: %s, Has Permission: %t\n",
			currentUser.ID.String(), targetUserID.String(), currentUser.Role, hasPermission)

		if !hasPermission {
			c.JSON(http.StatusForbidden, gin.H{"message": "No permission to modify username"})
			return
		}

		var req UpdateUsernameRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid username format"})
			return
		}

		// 检查用户名是否已存在 (排除目标用户)
		var count int64
		if err := db.Model(&models.User{}).Where("username = ? AND id != ?", req.Username, targetUserID).Count(&count).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
			return
		}

		if count > 0 {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Username already exists"})
			return
		}

		// 更新目标用户的用户名
		if err := db.Model(&models.User{}).Where("id = ?", targetUserID).Update("username", req.Username).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to update username"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Username updated successfully"})
	}
}

// HandleDeleteUser 处理删除用户的请求
func HandleDeleteUser(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从中间件获取当前用户ID
		currentUserIDInterface, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
			return
		}

		// 将当前用户ID interface{} 转换为 string 类型
		currentUserIDStr, ok := currentUserIDInterface.(string)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to get current user ID string"})
			return
		}

		// 将当前用户ID字符串转换为 uuid.UUID 类型
		currentUserID, err := uuid.Parse(currentUserIDStr)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to parse current user ID"})
			return
		}

		// 绑定请求体
		var req DeleteUserRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request format"})
			return
		}

		// 检查用户是否确认删除
		if !req.Confirm {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Please confirm account deletion"})
			return
		}

		// 查询用户信息
		var user models.User
		if err := database.DB.First(&user, "id = ?", currentUserID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"message": "User not found"})
			return
		}

		// 删除用户
		if err := database.DB.Delete(&user).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to delete user"})
			return
		}

		// 清除用户的 cookie
		c.SetCookie("token", "", -1, "/", "", false, true)

		c.JSON(http.StatusOK, gin.H{"message": "Account deleted successfully"})
	}
}
