package handlers

import (
	"go-fundraising/auth/models"
	"go-fundraising/auth/services"
	campaign "go-fundraising/campaign/services"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gocql/gocql"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var userService = services.UserService{}
var campaignService = campaign.CampaignService{}

func CreateUserHandler(c *gin.Context) {
	var request struct {
		Email      string `json:"email"`
		Username   string `json:"username"`
		Password   string `json:"password"`
		RePassword string `json:"re_password"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}
	if request.RePassword != request.Password {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Password and RePassword are not matching"})
		return
	}
	exists, err := userService.CheckUsernameExists(c, request.Username)
	if exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username already taken"})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Password hashing failed"})
		return
	}

	user := models.User{
		ID:           gocql.TimeUUID(),
		Username:     request.Username,
		Email:        request.Email,
		PasswordHash: string(hashedPassword),
		CreatedAt:    time.Now(),
	}
	if err := userService.NewUser(c, user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to register user"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "User created",
		"user":    user,
	})
}

func LoginHandler(c *gin.Context) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	user, err := userService.GetUserByUsername(c, req.Username)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}

	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)) != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}
	jwtKey := []byte(os.Getenv("JWT_SECRET"))
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID.String(),
		"exp":     time.Now().Add(2 * time.Hour).Unix(),
	})

	accessTokenString, err := accessToken.SignedString(jwtKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID.String(),
		"exp":     time.Now().Add(7 * 24 * time.Hour).Unix(),
	})
	refreshTokenString, _ := refreshToken.SignedString(jwtKey)
	refreshTokenExpiry := time.Now().Add(7 * 24 * time.Hour)
	err = userService.SaveRefreshToken(c, refreshTokenString, user.ID, refreshTokenExpiry)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save token"})
		return
	}

	c.JSON(200, gin.H{
		"access_token":  accessTokenString,
		"refresh_token": refreshTokenString,
	})

}

func RefreshHandler(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "invalid request"})
		return
	}

	userID, err := userService.ValidateRefreshToken(c, req.RefreshToken)
	if err != nil {
		c.JSON(401, gin.H{"error": "invalid or expired refresh token"})
		return
	}
	_ = userService.DeleteRefreshToken(c, req.RefreshToken)
	jwtKey := []byte(os.Getenv("JWT_SECRET"))
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID.String(),
		"exp":     time.Now().Add(2 * time.Hour).Unix(),
	})
	accessTokenString, _ := accessToken.SignedString(jwtKey)

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID.String(),
		"exp":     time.Now().Add(7 * 24 * time.Hour).Unix(),
	})
	refreshTokenString, _ := refreshToken.SignedString(jwtKey)
	refreshTokenExpiry := time.Now().Add(7 * 24 * time.Hour)
	err = userService.SaveRefreshToken(c, refreshTokenString, userID, refreshTokenExpiry)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save token"})
		return
	}

	c.JSON(200, gin.H{
		"access_token":  accessTokenString,
		"refresh_token": refreshTokenString,
	})
}

func LogoutHandler(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "refresh_token is required"})
		return
	}

	refreshToken := req.RefreshToken

	_, err := userService.ValidateRefreshToken(c, refreshToken)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"message": "logged out"})
		return
	}

	if err := userService.DeleteRefreshToken(c, refreshToken); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete refresh token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "logged out",
	})
}

func CurrentUserHandler(c *gin.Context) {
	row, exists := c.Get("user_id")
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

	userID := row.(gocql.UUID)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	user, err := userService.GetUserByID(c, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user"})
		return
	}
	listCampaign, err := campaignService.GetCampaignByUserID(c, user.ID.String(), page, perPage)

	c.JSON(http.StatusOK, gin.H{
		"id":         user.ID.String(),
		"email":      user.Email,
		"username":   user.Username,
		"created_at": user.CreatedAt,
		"campaigns":  listCampaign,
	})
}
