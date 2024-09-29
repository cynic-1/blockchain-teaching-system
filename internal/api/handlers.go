package api

import (
	"fmt"
	"github.com/cynic-1/blockchain-teaching-system/internal/auth"
	"github.com/cynic-1/blockchain-teaching-system/internal/database"
	"github.com/cynic-1/blockchain-teaching-system/internal/docker"
	"github.com/cynic-1/blockchain-teaching-system/internal/models"
	"github.com/gin-gonic/gin"
	"net/http"
)

type Handler struct {
	DB     database.DatabaseInterface
	Docker *docker.DockerManager
}

// httpError 用于封装 HTTP 错误
type httpError struct {
	StatusCode int
	Message    string
}

// Error 实现 error 接口
func (e *httpError) Error() string {
	return fmt.Sprintf("HTTP %d: %s", e.StatusCode, e.Message)
}

// getUserFromContext 从 gin.Context 中获取用户信息
func (h *Handler) getUserFromContext(c *gin.Context) (*models.User, error) {
	userIDInterface, exists := c.Get("userID")
	if !exists {
		return nil, &httpError{http.StatusBadRequest, "No userID"}
	}

	userID, ok := userIDInterface.(string)
	if !ok {
		return nil, &httpError{http.StatusInternalServerError, "UserID is not a string"}
	}

	user, err := h.DB.GetUser(userID)
	if err != nil {
		return nil, &httpError{http.StatusInternalServerError, "Failed to get user"}
	}

	return user, nil
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
	user, err := h.getUserFromContext(c)
	if err != nil {
		if httpErr, ok := err.(*httpError); ok {
			c.JSON(httpErr.StatusCode, gin.H{"error": httpErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unknown error"})
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
	user, err := h.getUserFromContext(c)
	if err != nil {
		if httpErr, ok := err.(*httpError); ok {
			c.JSON(httpErr.StatusCode, gin.H{"error": httpErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unknown error"})
		return
	}

	err = h.Docker.StartContainer(c.Request.Context(), user.ContainerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start container"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"Successfully started container": user.ContainerID})
}

func (h *Handler) StopContainer(c *gin.Context) {
	user, err := h.getUserFromContext(c)
	if err != nil {
		handleHttpError(c, err)
		return
	}

	err = h.Docker.StopContainer(c.Request.Context(), user.ContainerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"result": "Successfully Stopped Container"})
}

func (h *Handler) RemoveContainer(c *gin.Context) {
	user, err := h.getUserFromContext(c)
	if err != nil {
		handleHttpError(c, err)
		return
	}

	err = h.Docker.RemoveContainer(c.Request.Context(), user.ContainerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"result": "Successfully Removed Container"})
}

// 辅助函数，用于处理 HTTP 错误
func handleHttpError(c *gin.Context, err error) {
	if httpErr, ok := err.(*httpError); ok {
		c.JSON(httpErr.StatusCode, gin.H{"error": httpErr.Message})
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unknown error"})
	}
}

func (h *Handler) Exec(c *gin.Context) {
	var command struct {
		Cmd []string `json:"cmd"`
	}
	if err := c.ShouldBindJSON(&command); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len(command.Cmd) < 2 || command.Cmd[0] != "mis" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Command"})
	}

	user, err := h.getUserFromContext(c)
	if err != nil {
		if httpErr, ok := err.(*httpError); ok {
			c.JSON(httpErr.StatusCode, gin.H{"error": httpErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unknown error"})
		return
	}

	var body = ""
	if len(command.Cmd) > 2 {
		body = command.Cmd[2]
	}
	output, err := h.Docker.SendRequest(user.ContainerID, command.Cmd[1], body)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err, "output": output})
		return
	}
	c.JSON(http.StatusOK, gin.H{"output": output})
}

// 其他处理器方法...

// deprecated
//func (h *Handler) Exec(c *gin.Context) {
//	var command struct {
//		Cmd []string `json:"cmd"`
//	}
//	if err := c.ShouldBindJSON(&command); err != nil {
//		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
//		return
//	}
//
//	user, err := h.getUserFromContext(c)
//	if err != nil {
//		if httpErr, ok := err.(*httpError); ok {
//			c.JSON(httpErr.StatusCode, gin.H{"error": httpErr.Message})
//			return
//		}
//		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unknown error"})
//		return
//	}
//
//	output, err := h.Docker.ExecuteShellCommand(user.ContainerID, command.Cmd)
//	if err != nil {
//		c.JSON(http.StatusInternalServerError, gin.H{"error": err, "output": output})
//		return
//	}
//	c.JSON(http.StatusOK, gin.H{"output": output})
//}
//
//func (h *Handler) GetConsensusStatus(c *gin.Context) {
//	user, err := h.getUserFromContext(c)
//	if err != nil {
//		handleHttpError(c, err)
//		return
//	}
//
//	status, err := h.Docker.GetConsensusStatus(user.ContainerID)
//	if err != nil {
//		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
//		return
//	}
//	c.JSON(http.StatusOK, gin.H{"status": status})
//}
//
//func (h *Handler) GetTxpoolStatus(c *gin.Context) {
//	user, err := h.getUserFromContext(c)
//	if err != nil {
//		handleHttpError(c, err)
//		return
//	}
//
//	status, err := h.Docker.GetTxpoolStatus(user.ContainerID)
//	if err != nil {
//		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
//		return
//	}
//	c.JSON(http.StatusOK, gin.H{"status": status})
//}
//
//func (h *Handler) GetBlockAtHeight(c *gin.Context) {
//	user, err := h.getUserFromContext(c)
//	if err != nil {
//		handleHttpError(c, err)
//		return
//	}
//
//	heightStr := c.Query("height")
//	height, err := strconv.Atoi(heightStr)
//	if err != nil {
//		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid height"})
//		return
//	}
//
//	block, err := h.Docker.GetBlockAtHeight(user.ContainerID, height)
//	if err != nil {
//		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
//		return
//	}
//	c.JSON(http.StatusOK, gin.H{"block": block})
//}
//
//func (h *Handler) CreateLocalClusterFactory(c *gin.Context) {
//	user, err := h.getUserFromContext(c)
//	if err != nil {
//		handleHttpError(c, err)
//		return
//	}
//
//	var req struct {
//		NodeCount  int `json:"nodeCount"`
//		StakeQuota int `json:"stakeQuota"`
//		WindowSize int `json:"windowSize"`
//	}
//	if err := c.ShouldBindJSON(&req); err != nil {
//		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
//		return
//	}
//
//	result, err := h.Docker.CreateLocalClusterFactory(user.ContainerID, req.NodeCount, req.StakeQuota, req.WindowSize)
//	if err != nil {
//		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
//		return
//	}
//	c.JSON(http.StatusOK, gin.H{"result": result})
//}
//
//func (h *Handler) MakeLocalAddresses(c *gin.Context) {
//	user, err := h.getUserFromContext(c)
//	if err != nil {
//		handleHttpError(c, err)
//		return
//	}
//
//	result, err := h.Docker.MakeLocalAddresses(user.ContainerID)
//	if err != nil {
//		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
//		return
//	}
//	c.JSON(http.StatusOK, gin.H{"result": result})
//}
//
//func (h *Handler) MakeValidatorKeysAndStakeQuotas(c *gin.Context) {
//	user, err := h.getUserFromContext(c)
//	if err != nil {
//		handleHttpError(c, err)
//		return
//	}
//
//	result, err := h.Docker.MakeValidatorKeysAndStakeQuotas(user.ContainerID)
//	if err != nil {
//		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
//		return
//	}
//	c.JSON(http.StatusOK, gin.H{"result": result})
//}
//
//func (h *Handler) WriteGenesisFiles(c *gin.Context) {
//	user, err := h.getUserFromContext(c)
//	if err != nil {
//		handleHttpError(c, err)
//		return
//	}
//
//	result, err := h.Docker.WriteGenesisFiles(user.ContainerID)
//	if err != nil {
//		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
//		return
//	}
//	c.JSON(http.StatusOK, gin.H{"result": result})
//}
//
//func (h *Handler) CreateCluster(c *gin.Context) {
//	user, err := h.getUserFromContext(c)
//	if err != nil {
//		handleHttpError(c, err)
//		return
//	}
//
//	result, err := h.Docker.CreateCluster(user.ContainerID)
//	if err != nil {
//		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
//		return
//	}
//	c.JSON(http.StatusOK, gin.H{"result": result})
//}
//
//func (h *Handler) BuildBlockchainBinary(c *gin.Context) {
//	user, err := h.getUserFromContext(c)
//	if err != nil {
//		handleHttpError(c, err)
//		return
//	}
//
//	result, err := h.Docker.BuildBlockchainBinary(user.ContainerID)
//	if err != nil {
//		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
//		return
//	}
//	c.JSON(http.StatusOK, gin.H{"result": result})
//}
//
//func (h *Handler) ResetWorkingDirectory(c *gin.Context) {
//	user, err := h.getUserFromContext(c)
//	if err != nil {
//		handleHttpError(c, err)
//		return
//	}
//
//	result, err := h.Docker.ResetWorkingDirectory(user.ContainerID)
//	if err != nil {
//		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
//		return
//	}
//	c.JSON(http.StatusOK, gin.H{"result": result})
//}
//
//func (h *Handler) StartCluster(c *gin.Context) {
//	user, err := h.getUserFromContext(c)
//	if err != nil {
//		handleHttpError(c, err)
//		return
//	}
//
//	result, err := h.Docker.StartCluster(user.ContainerID)
//	if err != nil {
//		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
//		return
//	}
//	c.JSON(http.StatusOK, gin.H{"result": result})
//}
//
//func (h *Handler) StopCluster(c *gin.Context) {
//	user, err := h.getUserFromContext(c)
//	if err != nil {
//		handleHttpError(c, err)
//		return
//	}
//
//	result, err := h.Docker.StopCluster(user.ContainerID)
//	if err != nil {
//		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
//		return
//	}
//	c.JSON(http.StatusOK, gin.H{"result": result})
//}
