package api

import (
	"github.com/cynic-1/blockchain-teaching-system/internal/auth"
	"github.com/cynic-1/blockchain-teaching-system/internal/database"
	"github.com/cynic-1/blockchain-teaching-system/internal/docker"
	"github.com/cynic-1/blockchain-teaching-system/internal/models"
	"github.com/gin-gonic/gin"
	"net/http"
)

type Handler struct {
	DB     *database.Database
	Docker *docker.DockerManager
}

func (h *Handler) Register(c *gin.Context) {
	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hashedPassword, err := auth.HashPassword(user.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}
	user.Password = hashedPassword

	if err := h.DB.SaveUser(&user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User registered successfully"})
}

func (h *Handler) Login(c *gin.Context) {
	var loginData struct {
		UserID   string `json:"userID"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&loginData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.DB.GetUser(loginData.UserID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Could not get User form DB"})
		return
	}

	if !auth.ValidateUser(user, loginData.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	token, err := auth.CreateToken(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Create token errors"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"token":   token,
	})
}

func (h *Handler) CreateContainer(c *gin.Context) {
	//var containerConfig struct {
	//	UserID string   `json:"userID"`
	//	Image  string   `json:"image"`
	//	Cmd    []string `json:"cmd"`
	//}
	//if err := c.ShouldBindJSON(&containerConfig); err != nil {
	//	c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	//	return
	//}

	userIDInterface, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No userID"})
		return
	}

	userID, ok := userIDInterface.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "UserID is not a string"})
		return
	}

	user, err := h.DB.GetUser(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user"})
		return
	}

	containerID, err := h.Docker.CreateContainer(c.Request.Context(), "chain-proxy", nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create container"})
		return
	}

	user.ContainerID = containerID
	if err := h.DB.SaveUser(user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"Successfully created container": containerID})
}

func (h *Handler) StartContainer(c *gin.Context) {
	userIDInterface, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "No userID"})
		return
	}

	userID, ok := userIDInterface.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "UserID is not a string"})
		return
	}

	user, err := h.DB.GetUser(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user"})
		return
	}

	err = h.Docker.StartContainer(c.Request.Context(), user.ContainerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start container"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"Successfully started container": user.ContainerID})
}

func (h *Handler) Exec(c *gin.Context) {
	var command struct {
		Cmd []string `json:"cmd"`
	}
	if err := c.ShouldBindJSON(&command); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userIDInterface, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No userID"})
		return
	}

	userID, ok := userIDInterface.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "UserID is not a string"})
		return
	}

	user, err := h.DB.GetUser(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user"})
		return
	}

	output, err := h.Docker.ExecuteShellCommand(user.ContainerID, command.Cmd)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err, "output": output})
		return
	}
	c.JSON(http.StatusOK, gin.H{"output": output})
}

// 其他处理器方法...
