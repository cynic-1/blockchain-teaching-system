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
	// 创建 DockerManager
	dm, err := NewDockerManager("1.41") // 使用适合你的 Docker API 版本
	assert.NoError(t, err)

	ctx := context.Background()

	// 创建容器
	containerID, err := dm.CreateContainer(ctx, "chain-proxy", []string{"sleep", "infinity"})
	assert.NoError(t, err)
	assert.NotEmpty(t, containerID)

	// 确保在测试结束后清理容器
	defer func() {
		// 检查容器是否存在
		_, err := dm.client.ContainerInspect(ctx, containerID)
		if err == nil {
			err := dm.StopContainer(ctx, containerID)
			assert.NoError(t, err)
			err = dm.RemoveContainer(ctx, containerID)
			assert.NoError(t, err)
		}
	}()

	// 启动容器
	err = dm.StartContainer(ctx, containerID)
	assert.NoError(t, err)

	// 等待容器完全启动
	time.Sleep(2 * time.Second)

	// 在这里，你可以添加 ExecuteShellCommand 的测试
	// 例如：
	result, err := dm.ExecuteShellCommand(containerID, "echo 'Hello, World!'")
	assert.NoError(t, err)
	assert.Contains(t, string(result), "Hello, World!")

	// 测试停止容器
	err = dm.StopContainer(ctx, containerID)
	assert.NoError(t, err)

	// 测试删除容器
	err = dm.RemoveContainer(ctx, containerID)
	assert.NoError(t, err)
}
