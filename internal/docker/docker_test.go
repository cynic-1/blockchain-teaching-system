package docker

import (
	"context"
	"github.com/stretchr/testify/assert"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/docker/docker/api/types/container"
)

func TestDockerManager_HelloWorld(t *testing.T) {
	// 创建DockerManager实例
	dm, err := NewDockerManager("1.41") // 使用适合您环境的Docker API版本
	if err != nil {
		t.Fatalf("无法创建DockerManager: %v", err)
	}

	ctx := context.Background()

	// 创建hello-world容器
	containerID, err := dm.CreateContainer(ctx, "hello-world", nil)
	if err != nil {
		t.Fatalf("无法创建容器: %v", err)
	}

	// 启动容器
	err = dm.StartContainer(ctx, containerID)
	if err != nil {
		t.Fatalf("无法启动容器: %v", err)
	}

	// 等待容器运行完成
	statusCh, errCh := dm.client.ContainerWait(ctx, containerID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("等待容器时出错: %v", err)
		}
	case <-statusCh:
	case <-time.After(10 * time.Second):
		t.Fatal("等待容器超时")
	}

	// 获取容器日志
	out, err := dm.client.ContainerLogs(ctx, containerID, container.LogsOptions{ShowStdout: true})
	if err != nil {
		t.Fatalf("无法获取容器日志: %v", err)
	}
	defer out.Close()

	// 读取日志内容
	logContent, err := io.ReadAll(out)
	if err != nil {
		t.Fatalf("无法读取日志内容: %v", err)
	}

	// 检查日志内容
	if !strings.Contains(string(logContent), "Hello from Docker!") {
		t.Errorf("容器日志中未包含预期的输出")
	}

	t.Logf("容器日志: %s", logContent)

	// 删除容器
	err = dm.RemoveContainer(ctx, containerID)
	if err != nil {
		t.Fatalf("无法删除容器: %v", err)
	}
}

func TestDockerManager(t *testing.T) {
	dm, err := NewDockerManager("1.41") // 使用适合你的 Docker API 版本
	assert.NoError(t, err)

	ctx := context.Background()

	// 创建容器
	containerID, err := dm.CreateContainer(ctx, "chain-proxy", nil)
	assert.NoError(t, err)
	assert.NotEmpty(t, containerID)
	t.Log(containerID)

	// 立即检查容器是否成功创建
	_, err = dm.client.ContainerInspect(ctx, containerID)
	assert.NoError(t, err, "Container was not created successfully")

	// 清理函数
	defer func() {
		t.Log("Attempting to clean up container:", containerID)
		// 检查容器是否存在
		_, err := dm.client.ContainerInspect(ctx, containerID)
		if err == nil {
			t.Log("Container exists, stopping it")
			err := dm.StopContainer(ctx, containerID)
			if err != nil {
				t.Logf("Error stopping container: %v", err)
			}
			t.Log("Removing container")
			err = dm.RemoveContainer(ctx, containerID)
			if err != nil {
				t.Logf("Error removing container: %v", err)
			}
		} else {
			t.Log("Container does not exist, skipping cleanup")
		}
	}()

	// 启动容器
	t.Log("Starting container")
	err = dm.StartContainer(ctx, containerID)
	assert.NoError(t, err)

	// 等待容器完全启动
	time.Sleep(2 * time.Second)

	// 检查容器是否仍然存在
	t.Log("Checking if container still exists")
	_, err = dm.client.ContainerInspect(ctx, containerID)
	if err != nil {
		t.Fatalf("Container no longer exists after starting: %v", err)
	}

	// 执行命令
	//t.Log("Executing command in container")
	//result, err := dm.ExecuteShellCommand(containerID, "echo Hello, World!")
	//if err != nil {
	//	t.Logf("Error executing command: %v", err)
	//} else {
	//	assert.Contains(t, string(result), "Hello, World!")
	//}

	result, err := dm.CreateLocalClusterFactory(containerID, 4, 9999, 4)
	if err != nil {
		t.Logf("Error creating local cluster: %v", err)
	} else {
		t.Log(result)
	}

	result, err = dm.ResetWorkingDirectory(containerID)
	if err != nil {
		t.Logf("Error reseting workDir: %v", err)
	} else {
		t.Log(result)
	}

	result, err = dm.ExecuteShellCommand(containerID, []string{"ls", "-l"})
	if err != nil {
		t.Logf("Error executing command: %v", err)
	} else {
		t.Log(result)
	}

	result, err = dm.MakeValidatorKeysAndStakeQuotas(containerID)
	if err != nil {
		t.Logf("Error generating validator keys and stake quotas: %v", err)
	} else {
		t.Log(result)
	}

	result, err = dm.MakeLocalAddresses(containerID)
	if err != nil {
		t.Logf("Error making local addresses: %v", err)
	} else {
		t.Log(result)
	}

	result, err = dm.WriteGenesisFiles(containerID)
	if err != nil {
		t.Logf("Error writing genesis files: %v", err)
	} else {
		t.Log(result)
	}

	result, err = dm.BuildBlockchainBinary(containerID)
	if err != nil {
		t.Logf("Error building blockchain binary: %v", err)
	} else {
		t.Log(result)
	}

	result, err = dm.CreateCluster(containerID)
	if err != nil {
		t.Logf("Error creating new cluster: %v", err)
	} else {
		t.Log(result)
	}

	result, err = dm.StartCluster(containerID)
	if err != nil {
		t.Logf("Error starting cluster: %v", err)
	} else {
		t.Log(result)
	}

	result, err = dm.GetConsensusStatus(containerID)
	if err != nil {
		t.Logf("Error getting consensus status: %v", err)
	} else {
		t.Log(result)
	}

	time.Sleep(5 * time.Second)

	result, err = dm.GetConsensusStatus(containerID)
	if err != nil {
		t.Logf("Error getting consensus status: %v", err)
	} else {
		t.Log(result)
	}

	result, err = dm.GetTxpoolStatus(containerID)
	if err != nil {
		t.Logf("Error getting txpool status: %v", err)
	} else {
		t.Log(result)
	}

	result, err = dm.StopCluster(containerID)
	if err != nil {
		t.Logf("Error stoping cluster: %v", err)
	} else {
		t.Log(result)
	}

	// 停止容器
	t.Log("Stopping container")
	err = dm.StopContainer(ctx, containerID)
	if err != nil {
		t.Logf("Error stopping container: %v", err)
	}

	// 删除容器
	t.Log("Removing container")
	err = dm.RemoveContainer(ctx, containerID)
	if err != nil {
		t.Logf("Error removing container: %v", err)
	}
}
