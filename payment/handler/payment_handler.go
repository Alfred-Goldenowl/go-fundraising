package handlers

import (
	"context"
	auth "go-fundraising/auth/services"
	campaign "go-fundraising/campaign/services"
	"go-fundraising/payment/models"
	payment "go-fundraising/payment/services"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gocql/gocql"
	"github.com/stripe/stripe-go/v74"
	"github.com/stripe/stripe-go/v74/checkout/session"
)

var userService = auth.UserService{}
var paymentService = payment.PaymentService{}
var campaignService = campaign.CampaignService{}

type CreatePaymentIntentRequest struct {
	Amount     int64  `json:"amount" binding:"required"`
	Currency   string `json:"currency"`
	CampaignID string `json:"campaign_id"`
}

type PaymentSuccessData struct {
	CheckoutID string
	UserID     string
	CampaignID string
	Amount     int64
	Currency   string
	Status     string
}

func CreatePaymentIntentHandler(c *gin.Context) {
	var req CreatePaymentIntentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	raw, _ := c.Get("user_id")
	userID := raw.(gocql.UUID)

	user, _ := userService.GetUserByID(context.Background(), userID)
	if user.ID == (gocql.UUID{}) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	if req.Currency == "" {
		req.Currency = "usd"
	}

	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")
	host := os.Getenv("APP_HOST")

	successURL := host + "/payment/success" + "?session_id={CHECKOUT_SESSION_ID}"
	failURL := host + "/payment/fail" + "?session_id={CHECKOUT_SESSION_ID}"

	params := &stripe.CheckoutSessionParams{
		PaymentMethodTypes: stripe.StringSlice([]string{"card"}),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
					Currency: stripe.String(req.Currency),
					ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
						Name: stripe.String("Donation for campaign " + req.CampaignID),
					},
					UnitAmount: stripe.Int64(req.Amount * 100),
				},
				Quantity: stripe.Int64(1),
			},
		},
		Mode:       stripe.String(string(stripe.CheckoutSessionModePayment)),
		SuccessURL: stripe.String(successURL),
		CancelURL:  stripe.String(failURL),
	}
	params.Metadata = map[string]string{
		"user_id":     user.ID.String(),
		"campaign_id": req.CampaignID,
	}

	s, err := session.New(params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"url": s.URL,
	})
}

func PaymentSuccessHandler(c *gin.Context) {
	checkoutID := c.Query("session_id")
	if checkoutID == "" {
		c.String(http.StatusBadRequest, "session_id is required")
		return
	}

	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")

	sess, err := session.Get(checkoutID, nil)
	if err != nil {
		c.String(http.StatusInternalServerError, "failed to get checkout session")
		return
	}

	isExist, err := paymentService.CheckoutExists(checkoutID)
	if !isExist {
		CampaignID, _ := gocql.ParseUUID(sess.Metadata["campaign_id"])
		UserID, _ := gocql.ParseUUID(sess.Metadata["user_id"])
		user, _ := userService.GetUserByID(context.Background(), UserID)
		currentPayment := models.PaymentHistory{
			ID:         gocql.TimeUUID(),
			CampaignID: CampaignID,
			Username:   user.Username,
			UserID:     UserID,
			CreatedAt:  time.Now(),
			CheckoutID: checkoutID,
			Amount:     sess.AmountTotal / 100,
		}

		if err := paymentService.NewPayment(context.Background(), currentPayment); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create currentPayment"})
			return
		}

		if err := campaignService.UpdateCampaignAmountCollected(context.Background(), CampaignID, sess.AmountTotal/100); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	data := PaymentSuccessData{
		CheckoutID: sess.ID,
		UserID:     sess.Metadata["user_id"],
		CampaignID: sess.Metadata["campaign_id"],
		Amount:     sess.AmountTotal / 10,
		Currency:   string(sess.Currency),
		Status:     string(sess.PaymentStatus),
	}

	c.HTML(http.StatusOK, "success.html", data)
}

func PaymentFailHandler(c *gin.Context) {
	checkoutID := c.Query("session_id")
	if checkoutID == "" {
		c.String(http.StatusBadRequest, "session_id is required")
		return
	}

	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")

	sess, err := session.Get(checkoutID, nil)
	if err != nil {
		c.String(http.StatusInternalServerError, "failed to get checkout session")
		return
	}

	data := PaymentSuccessData{
		CheckoutID: checkoutID,
		UserID:     sess.Metadata["user_id"],
		CampaignID: sess.Metadata["campaign_id"],
		Amount:     sess.AmountTotal,
		Currency:   string(sess.Currency),
		Status:     string(sess.PaymentStatus),
	}

	c.HTML(http.StatusOK, "fail.html", data)
}
