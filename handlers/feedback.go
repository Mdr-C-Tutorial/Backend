package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"mdr/utils"
)

type FeedbackRequest struct {
	Content string `json:"content" binding:"required"`
}

func HandleFeedback(c *gin.Context) {
	var req FeedbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid feedback content"})
		return
	}

	// 获取当前用户信息
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}

	// 发送反馈邮件
	err := utils.SendFeedbackEmail("295871597@qq.com", req.Content, userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to send feedback"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Feedback sent successfully"})
}