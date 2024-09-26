package api_test

import (
	"bytes"
	"encoding/json"
	"github.com/cynic-1/blockchain-teaching-system/internal/auth"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/cynic-1/blockchain-teaching-system/internal/api"
	"github.com/cynic-1/blockchain-teaching-system/internal/database"
	"github.com/cynic-1/blockchain-teaching-system/internal/docker"
)

func TestBlockchainOperations(t *testing.T) {
	// 设置测试环境
	// 设置测试环境
	gin.SetMode(gin.TestMode)
	router := gin.Default()

	err := auth.InitSecretKey()
	if err != nil {
		return
	}

	// 初始化真实的数据库和 Docker 管理器
	db, err := database.NewDatabase("test_db_connection_string")
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	dockerManager, err := docker.NewDockerManager("1.41")
	if err != nil {
		t.Fatalf("Failed to create Docker manager: %v", err)
	}

	// 创建 Handler 并设置路由
	handler := &api.Handler{
		DB:     db,
		Docker: dockerManager,
	}
	// 公开路由组，不需要 token 验证
	public := router.Group("/api")
	{
		public.POST("/register", handler.Register)
		public.POST("/login", handler.Login)
	}

	// 受保护的路由组，需要 token 验证
	protected := router.Group("/api")
	protected.Use(auth.JWTMiddleware())
	{
		protected.POST("/container/create", handler.CreateContainer)
		protected.POST("/container/start", handler.StartContainer)
		protected.POST("/container/exec", handler.Exec)
		protected.POST("/container/stop", handler.StopContainer)
		protected.POST("/container/remove", handler.RemoveContainer)
		protected.GET("/consensus-status", handler.GetConsensusStatus)
		protected.GET("/txpool-status", handler.GetTxpoolStatus)
		protected.GET("/block", handler.GetBlockAtHeight)
		protected.POST("/create-cluster-factory", handler.CreateLocalClusterFactory)
		protected.POST("/make-local-addresses", handler.MakeLocalAddresses)
		protected.POST("/make-validator-keys", handler.MakeValidatorKeysAndStakeQuotas)
		protected.POST("/write-genesis-files", handler.WriteGenesisFiles)
		protected.POST("/build-blockchain", handler.BuildBlockchainBinary)
		protected.POST("/reset-working-directory", handler.ResetWorkingDirectory)
		protected.POST("/cluster/create", handler.CreateCluster)
		protected.POST("/cluster/start", handler.StartCluster)
		protected.POST("/cluster/stop", handler.StopCluster)
		// 添加其他需要验证的路由...
	}

	// 测试用户信息
	userID := "testuser"
	password := "testpassword"

	// 1. 测试注册
	t.Run("Register", func(t *testing.T) {
		body := map[string]string{
			"userID":   userID,
			"password": password,
		}
		jsonBody, _ := json.Marshal(body)
		req, _ := http.NewRequest("POST", "/api/register", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "User registered successfully")

		retrievedUser, err := db.GetUser(userID)
		assert.NoError(t, err, "Should be able to retrieve the registered user")
		assert.NotNil(t, retrievedUser, "Retrieved user should not be nil")
		assert.Equal(t, userID, retrievedUser.ID, "Retrieved user ID should match")
		t.Logf("Retrieved user password hash: %s", retrievedUser.Password)
	})

	time.Sleep(time.Millisecond * 100) // 在注册和登录之间添加小延迟
	// 2. 测试登录
	var token string
	t.Run("Login", func(t *testing.T) {
		body := map[string]string{
			"userID":   userID,
			"password": password,
		}
		jsonBody, _ := json.Marshal(body)

		t.Logf("Login request body: %s", string(jsonBody))

		retrievedUser, err := db.GetUser(userID)
		assert.NoError(t, err, "Should be able to retrieve the registered user")
		assert.NotNil(t, retrievedUser, "Retrieved user should not be nil")
		assert.Equal(t, userID, retrievedUser.ID, "Retrieved user ID should match")
		t.Logf("Retrieved user password hash: %s", retrievedUser.Password)

		req, _ := http.NewRequest("POST", "/api/login", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		t.Logf(w.Body.String())

		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]string
		err = json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			return
		}
		token = response["token"]
		assert.NotEmpty(t, token)
	})

	// 启动容器
	t.Run("CreateContainer", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/api/container/create", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	// 启动容器
	t.Run("StartContainer", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/api/container/start", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	// 等待容器完全启动
	time.Sleep(2 * time.Second)

	// 执行命令
	t.Run("ExecuteCommand", func(t *testing.T) {
		body := map[string]interface{}{
			"cmd": []string{"echo", "Hello, World!"},
		}
		jsonBody, _ := json.Marshal(body)
		req, _ := http.NewRequest("POST", "/api/container/exec", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]string
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Contains(t, response["output"], "Hello, World!")
	})

	// 创建本地集群工厂
	t.Run("CreateLocalClusterFactory", func(t *testing.T) {
		body := map[string]int{
			"nodeCount":  4,
			"stakeQuota": 9999,
			"windowSize": 4,
		}
		jsonBody, _ := json.Marshal(body)
		req, _ := http.NewRequest("POST", "/api/create-cluster-factory", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	// 重置工作目录
	t.Run("ResetWorkingDirectory", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/api/reset-working-directory", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	// 执行ls -l命令
	t.Run("ExecuteLsCommand", func(t *testing.T) {
		body := map[string]interface{}{
			"cmd": []string{"ls", "-l"},
		}
		jsonBody, _ := json.Marshal(body)
		req, _ := http.NewRequest("POST", "/api/container/exec", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	// 创建验证者密钥和权益配额
	t.Run("MakeValidatorKeysAndStakeQuotas", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/api/make-validator-keys", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	// 创建本地地址
	t.Run("MakeLocalAddresses", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/api/make-local-addresses", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	// 写入创世文件
	t.Run("WriteGenesisFiles", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/api/write-genesis-files", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	// 构建区块链二进制文件
	t.Run("BuildBlockchainBinary", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/api/build-blockchain", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	// 创建集群
	t.Run("CreateCluster", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/api/cluster/start", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	// 启动集群
	t.Run("StartCluster", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/api/cluster/start", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	// 获取共识状态
	t.Run("GetConsensusStatus", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/consensus-status", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	time.Sleep(5 * time.Second)

	// 再次获取共识状态
	t.Run("GetConsensusStatusAgain", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/consensus-status", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	// 获取交易池状态
	t.Run("GetTxpoolStatus", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/txpool-status", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	// 停止集群
	t.Run("StopCluster", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/api/cluster/stop", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	// 停止容器
	t.Run("StopContainer", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/api/container/stop", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	// 删除容器
	t.Run("RemoveContainer", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/api/container/remove", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestIntegration(t *testing.T) {
	// 设置测试环境
	gin.SetMode(gin.TestMode)
	router := gin.Default()

	err := auth.InitSecretKey()
	if err != nil {
		return
	}

	// 初始化真实的数据库和 Docker 管理器
	db, err := database.NewDatabase("test_db_connection_string")
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	dockerManager, err := docker.NewDockerManager("1.41")
	if err != nil {
		t.Fatalf("Failed to create Docker manager: %v", err)
	}

	// 创建 Handler 并设置路由
	handler := &api.Handler{
		DB:     db,
		Docker: dockerManager,
	}
	// 公开路由组，不需要 token 验证
	public := router.Group("/api")
	{
		public.POST("/register", handler.Register)
		public.POST("/login", handler.Login)
	}

	// 受保护的路由组，需要 token 验证
	protected := router.Group("/api")
	protected.Use(auth.JWTMiddleware())
	{
		protected.POST("/container/create", handler.CreateContainer)
		protected.POST("/container/start", handler.StartContainer)
		protected.POST("/container/exec", handler.Exec)
		// 添加其他需要验证的路由...
	}

	//router.POST("/register", handler.Register)
	//router.POST("/login", handler.Login)
	//router.POST("/createContainer", handler.CreateContainer)
	//router.POST("/exec", handler.Exec)

	// 测试用户信息
	userID := "testuser"
	password := "testpassword"

	// 1. 测试注册
	t.Run("Register", func(t *testing.T) {
		body := map[string]string{
			"userID":   userID,
			"password": password,
		}
		jsonBody, _ := json.Marshal(body)
		req, _ := http.NewRequest("POST", "/api/register", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "User registered successfully")

		retrievedUser, err := db.GetUser(userID)
		assert.NoError(t, err, "Should be able to retrieve the registered user")
		assert.NotNil(t, retrievedUser, "Retrieved user should not be nil")
		assert.Equal(t, userID, retrievedUser.ID, "Retrieved user ID should match")
		t.Logf("Retrieved user password hash: %s", retrievedUser.Password)
	})

	time.Sleep(time.Millisecond * 100) // 在注册和登录之间添加小延迟
	// 2. 测试登录
	var token string
	t.Run("Login", func(t *testing.T) {
		body := map[string]string{
			"userID":   userID,
			"password": password,
		}
		jsonBody, _ := json.Marshal(body)

		t.Logf("Login request body: %s", string(jsonBody))

		retrievedUser, err := db.GetUser(userID)
		assert.NoError(t, err, "Should be able to retrieve the registered user")
		assert.NotNil(t, retrievedUser, "Retrieved user should not be nil")
		assert.Equal(t, userID, retrievedUser.ID, "Retrieved user ID should match")
		t.Logf("Retrieved user password hash: %s", retrievedUser.Password)

		req, _ := http.NewRequest("POST", "/api/login", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		t.Logf(w.Body.String())

		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]string
		err = json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			return
		}
		token = response["token"]
		assert.NotEmpty(t, token)
	})

	// 3. 测试创建容器
	t.Run("CreateContainer", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/api/container/create", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "Successfully created container")
	})

	// 4. 测试创建容器
	t.Run("StartContainer", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/api/container/start", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "Successfully started container")
	})

	// 5. 测试执行命令
	t.Run("Exec", func(t *testing.T) {
		body := map[string]interface{}{
			"cmd": []string{"ls", "-l"},
		}
		jsonBody, _ := json.Marshal(body)

		t.Logf("Exec request body: %s", string(jsonBody))

		req, _ := http.NewRequest("POST", "/api/container/exec", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		t.Logf(w.Body.String())
		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]string
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.NotEmpty(t, response["output"])
	})

	// 测试获取共识状态
	t.Run("GetConsensusStatus", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/consensus-status", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]string
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.NotEmpty(t, response["status"])
	})

	// 测试获取交易池状态
	t.Run("GetTxpoolStatus", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/txpool-status", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]string
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.NotEmpty(t, response["status"])
	})

	// 测试获取特定高度的区块
	t.Run("GetBlockAtHeight", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/block?height=1", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]string
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.NotEmpty(t, response["block"])
	})

	// 测试创建本地集群工厂
	t.Run("CreateLocalClusterFactory", func(t *testing.T) {
		body := map[string]int{
			"nodeCount":  4,
			"stakeQuota": 1000,
			"windowSize": 100,
		}
		jsonBody, _ := json.Marshal(body)

		req, _ := http.NewRequest("POST", "/api/create-cluster-factory", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]string
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.NotEmpty(t, response["result"])
	})

	// 测试创建本地点和主题地址
	t.Run("MakeLocalAddresses", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/api/make-local-addresses", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]string
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.NotEmpty(t, response["result"])
	})

	// 测试创建验证者密钥和权益配额
	t.Run("MakeValidatorKeysAndStakeQuotas", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/api/make-validator-keys", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]string
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.NotEmpty(t, response["result"])
	})

	// 测试写入创世文件
	t.Run("WriteGenesisFiles", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/api/write-genesis-files", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]string
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.NotEmpty(t, response["result"])
	})

	// 测试创建名为cluster_template的集群
	t.Run("CreateCluster", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/api/create-cluster", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]string
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.NotEmpty(t, response["result"])
	})

	// 测试构建区块链二进制文件
	t.Run("BuildBlockchainBinary", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/api/build-blockchain", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]string
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.NotEmpty(t, response["result"])
	})

	// 测试重置工作目录
	t.Run("ResetWorkingDirectory", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/api/reset-working-directory", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]string
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.NotEmpty(t, response["result"])
	})

	// 测试启动集群
	t.Run("StartCluster", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/api/start-cluster", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]string
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.NotEmpty(t, response["result"])
	})

	// 测试停止集群
	t.Run("StopCluster", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/api/stop-cluster", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]string
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.NotEmpty(t, response["result"])
	})
}
