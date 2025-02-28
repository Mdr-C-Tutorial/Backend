package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"mdr/models"
)

type UpdateUsernameRequest struct {
	Username string `json:"username" binding:"required,min=3,max=20"`
}

func UpdateUsername(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        userID, err := uuid.Parse(c.Param("id"))
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid user ID"})
            return
        }

        // 从中间件获取当前用户ID
        currentUserID, exists := c.Get("userID")
        if !exists {
            c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
            return
        }

        // 查询当前用户信息
        var currentUser models.User
        if err := db.First(&currentUser, "id = ?", currentUserID).Error; err != nil {
            c.JSON(http.StatusUnauthorized, gin.H{"message": "User not found"})
            return
        }

        // 验证权限
        if currentUser.ID != userID && currentUser.Role != "admin" {
            c.JSON(http.StatusForbidden, gin.H{"message": "No permission to modify username"})
            return
        }

		var req UpdateUsernameRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid username format"})
			return
		}

		// 检查用户名是否已存在
		var count int64
		if err := db.Model(&models.User{}).Where("username = ? AND id != ?", req.Username, userID).Count(&count).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
			return
		}

		if count > 0 {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Username already exists"})
			return
		}

		// 更新用户名
		if err := db.Model(&models.User{}).Where("id = ?", userID).Update("username", req.Username).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to update username"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Username updated successfully"})
	}
}