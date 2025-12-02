package handlers

import (
	"context"
	auth "go-fundraising/auth/services"
	"go-fundraising/campaign/models"
	campaign "go-fundraising/campaign/services"
	models2 "go-fundraising/payment/models"
	payment "go-fundraising/payment/services"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gocql/gocql"
)

var userService = auth.UserService{}
var campaignService = campaign.CampaignService{}
var paymentService = payment.PaymentService{}

type CampaignWithPayments struct {
	ID              gocql.UUID               `json:"ID"`
	Title           string                   `json:"Title"`
	Description     string                   `json:"Description"`
	Target          int                      `json:"Target"`
	AmountCollected int                      `json:"AmountCollected"`
	Image           string                   `json:"Image"`
	Deadline        time.Time                `json:"Deadline"`
	CreatedAt       time.Time                `json:"CreatedAt"`
	Payments        []models2.PaymentHistory `json:"Payments"`
}

func CreateCampaignHandler(c *gin.Context) {
	var request struct {
		Title       string    `json:"title"`
		Description string    `json:"description"`
		Target      int       `json:"target"`
		Image       string    `json:"image"`
		Deadline    time.Time `json:"deadline"`
	}

	raw, _ := c.Get("user_id")
	userID := raw.(gocql.UUID)

	user, _ := userService.GetUserByID(context.Background(), userID)

	if user.ID == (gocql.UUID{}) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}
	log.Println(user)
	CurrentCampaign := models.Campaign{
		ID:              gocql.TimeUUID(),
		UserID:          userID,
		Username:        user.Username,
		Title:           request.Title,
		Description:     request.Description,
		Target:          request.Target,
		Image:           request.Image,
		AmountCollected: 0,
		Deadline:        request.Deadline,
		CreatedAt:       time.Now(),
	}
	if _, err := campaignService.CreateCampaign(context.Background(), CurrentCampaign); err != nil {
		log.Print(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert campaign"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":  "Campaign successfully created",
		"campaign": CurrentCampaign,
	})
}

func GetCampaignHandler(c *gin.Context) {
	idParam := c.Param("campaign_id")
	if idParam == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
		return
	}

	campaignID, err := gocql.ParseUUID(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid UUID format"})
		return
	}

	campaign, err := campaignService.GetCampaignByID(c, campaignID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "campaign not found"})
		return
	}

	payments, err := paymentService.GetPaymentsByCampaignID(c, campaignID)
	if err != nil {
		payments = []models2.PaymentHistory{}
	}
	log.Println(payments)

	resp := CampaignWithPayments{
		ID:              campaign.ID,
		Title:           campaign.Title,
		Description:     campaign.Description,
		Target:          campaign.Target,
		AmountCollected: campaign.AmountCollected,
		Image:           campaign.Image,
		Deadline:        campaign.Deadline,
		CreatedAt:       campaign.CreatedAt,
		Payments:        payments,
	}

	c.JSON(http.StatusOK, resp)
}

func SearchCampaignHandler(c *gin.Context) {
	keyword := c.Query("q")
	pageStr := c.DefaultQuery("page", "1")
	perPageStr := c.DefaultQuery("per_page", "10")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}
	perPage, err := strconv.Atoi(perPageStr)
	if err != nil || perPage < 1 {
		perPage = 10
	}

	result, err := campaignService.SearchCampaign(keyword, page, perPage)
	if err != nil {
		log.Println("❌ Search error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to search campaigns"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"total":    result.Total,
		"data":     result.Data,
		"page":     page,
		"per_page": perPage,
	})
}
