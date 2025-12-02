package handlers

import (
	"context"
	"go-fundraising/campaign/models"
	"go-fundraising/campaign/services"
	"log"
	"net/http"
	"strconv"
	"time"

	"go-fundraising/configs"

	"github.com/gin-gonic/gin"
	"github.com/gocql/gocql"
)

var commentService = services.CommentService{}

func CreateCommentHandler(c *gin.Context) {
	var request struct {
		Content string `json:"content"`
	}
	CampaignIDString := c.Param("campaign_id")
	CampaignID, _ := gocql.ParseUUID(CampaignIDString)

	row, exists := c.Get("user_id")
	userID := row.(gocql.UUID)

	user, _ := userService.GetUserByID(context.Background(), userID)

	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	comment := models.Comment{
		ID:         gocql.TimeUUID(),
		UserID:     userID,
		CampaignID: CampaignID,
		Content:    request.Content,
		Username:   user.Username,
		CreatedAt:  time.Now(),
	}

	if err := commentService.InsertComment(context.Background(), comment); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert comment"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Comment successfully created",
		"user":    comment,
	})
}

func GetCommentsByCampaignIDHandler(c *gin.Context) {
	CampaignID := c.Param("campaign_id")
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", strconv.Itoa(configs.DefaultItemPerPage)))

	lastCreatedAtStr := c.Query("last_created_at")
	var lastCreatedAt time.Time
	if lastCreatedAtStr != "" {
		parsedTime, err := time.Parse(time.RFC3339, lastCreatedAtStr)
		if err == nil {
			lastCreatedAt = parsedTime
		}
	}

	comments, err := commentService.GetCommentsByCampaignID(context.Background(), CampaignID, perPage, lastCreatedAt)
	if err != nil {
		log.Print(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch comments"})
		return
	}

	c.JSON(http.StatusOK, comments)
}
