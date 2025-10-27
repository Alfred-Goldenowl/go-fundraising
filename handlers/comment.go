package handlers

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gocql/gocql"
	"go-fundraising/configs"
	"go-fundraising/models"
	"go-fundraising/services"
)

var commentService = services.CommentService{}

func CreateCommentHandler(c *gin.Context) {
	var request struct {
		UserID  string `json:"user_id"`
		PostID  string `json:"post_id"`
		Content string `json:"content"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	if request.UserID == "" || request.PostID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "UserID and PostID are required"})
		return
	}

	comment := models.Comment{
		ID:        gocql.TimeUUID(),
		UserID:    request.UserID,
		PostID:    request.PostID,
		Content:   request.Content,
		CreatedAt: time.Now(),
	}

	if err := commentService.InsertComment(context.Background(), comment); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert comment"})
		return
	}

	c.JSON(http.StatusCreated, comment)
}

func GetCommentsByPostIDHandler(c *gin.Context) {
	postID := c.Param("post_id")
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", strconv.Itoa(configs.DefaultItemPerPage)))

	lastCreatedAtStr := c.Query("last_created_at")
	var lastCreatedAt time.Time
	if lastCreatedAtStr != "" {
		parsedTime, err := time.Parse(time.RFC3339, lastCreatedAtStr)
		if err == nil {
			lastCreatedAt = parsedTime
		}
	}

	comments, err := commentService.GetCommentsByPostID(context.Background(), postID, perPage, lastCreatedAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch comments"})
		return
	}

	c.JSON(http.StatusOK, comments)
}
