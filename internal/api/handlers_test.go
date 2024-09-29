package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cynic-1/blockchain-teaching-system/internal/auth"
	"github.com/cynic-1/blockchain-teaching-system/internal/database"
	"github.com/cynic-1/blockchain-teaching-system/internal/docker"
	"github.com/cynic-1/blockchain-teaching-system/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupTestHandler() *Handler {
	mockDB := database.NewMockDatabase()
	mockDocker := &docker.DockerManager{}
	return &Handler{
		DB:     mockDB,
		Docker: mockDocker,
	}
}

func TestRegister(t *testing.T) {
	handler := setupTestHandler()
	router := gin.Default()
	router.POST("/register", handler.Register)

	user := models.User{
		ID:       "testuser",
		Password: "testpassword",
	}
	userJSON, _ := json.Marshal(user)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/register", bytes.NewBuffer(userJSON))
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "User registered successfully")
}

func TestLogin(t *testing.T) {
	handler := setupTestHandler()
	router := gin.Default()
	router.POST("/login", handler.Login)

	// 先注册一个用户
	user := &models.User{
		ID:       "testuser",
		Password: "testpassword",
	}
	hashedPassword, _ := auth.HashPassword(user.Password)
	user.Password = hashedPassword
	handler.DB.(*database.MockDatabase).SaveUser(user)

	loginData := map[string]string{
		"userID":   "testuser",
		"password": "testpassword",
	}
	loginJSON, _ := json.Marshal(loginData)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(loginJSON))
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Login successful")
	assert.Contains(t, w.Body.String(), "token")
}

func TestCreateContainer(t *testing.T) {
	handler := setupTestHandler()
	router := gin.Default()
	router.Use(func(c *gin.Context) {
		c.Set("userID", "testuser")
	})
	router.POST("/create", handler.CreateContainer)

	// 先添加一个用户
	user := &models.User{ID: "testuser"}
	handler.DB.(*database.MockDatabase).SaveUser(user)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/create", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Successfully created container")
}

func TestStartContainer(t *testing.T) {
	handler := setupTestHandler()
	router := gin.Default()
	router.Use(func(c *gin.Context) {
		c.Set("userID", "testuser")
	})
	router.POST("/start", handler.StartContainer)

	// 先添加一个用户，并设置 ContainerID
	user := &models.User{ID: "testuser", ContainerID: "test-container-id"}
	handler.DB.(*database.MockDatabase).SaveUser(user)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/start", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Successfully started container")
}

func TestStopContainer(t *testing.T) {
	handler := setupTestHandler()
	router := gin.Default()
	router.Use(func(c *gin.Context) {
		c.Set("userID", "testuser")
	})
	router.POST("/stop", handler.StopContainer)

	// 先添加一个用户，并设置 ContainerID
	user := &models.User{ID: "testuser", ContainerID: "test-container-id"}
	handler.DB.(*database.MockDatabase).SaveUser(user)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/stop", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Successfully Stopped Container")
}

func TestRemoveContainer(t *testing.T) {
	handler := setupTestHandler()
	router := gin.Default()
	router.Use(func(c *gin.Context) {
		c.Set("userID", "testuser")
	})
	router.POST("/remove", handler.RemoveContainer)

	// 先添加一个用户，并设置 ContainerID
	user := &models.User{ID: "testuser", ContainerID: "test-container-id"}
	handler.DB.(*database.MockDatabase).SaveUser(user)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/remove", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Successfully Removed Container")
}

func TestExec(t *testing.T) {
	handler := setupTestHandler()
	router := gin.Default()
	router.Use(func(c *gin.Context) {
		c.Set("userID", "testuser")
	})
	router.POST("/exec", handler.Exec)

	// 先添加一个用户，并设置 ContainerID
	user := &models.User{ID: "testuser", ContainerID: "test-container-id"}
	handler.DB.(*database.MockDatabase).SaveUser(user)

	command := map[string][]string{
		"cmd": {"mis", "test", "body"},
	}
	commandJSON, _ := json.Marshal(command)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/exec", bytes.NewBuffer(commandJSON))
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "output")
}
