package docker

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/pkg/stdcopy"
	"net/url"
	"strings"
)

func (dm *DockerManager) sendRequest(containerID, method, path string, body interface{}) (string, error) {
	var reqBody []byte
	var err error

	if body != nil {
		reqBody, err = json.Marshal(body)
		if err != nil {
			return "", err
		}
	}

	// 预设
	containerPort := "8080"

	// 创建执行命令
	cmd := []string{"curl", "-X", method, "-H", "Content-Type: application/json"}
	if body != nil {
		cmd = append(cmd, "-d", string(reqBody))
	}
	cmd = append(cmd, fmt.Sprintf("http://localhost:%s%s", containerPort, path))

	// 使用Docker客户端执行命令
	execConfig := container.ExecOptions{
		Cmd:          cmd,
		AttachStdout: true,
		AttachStderr: true,
	}

	ctx := context.Background()
	execID, err := dm.client.ContainerExecCreate(ctx, containerID, execConfig)
	if err != nil {
		return "", err
	}

	resp, err := dm.client.ContainerExecAttach(ctx, execID.ID, container.ExecStartOptions{})
	if err != nil {
		return "", err
	}
	defer resp.Close()

	// 使用 stdcopy.StdCopy 来分离 STDOUT 和 STDERR
	var outBuf, errBuf bytes.Buffer
	_, err = stdcopy.StdCopy(&outBuf, &errBuf, resp.Reader)
	if err != nil {
		return string([]byte("")), fmt.Errorf("error reading exec output: %v", err)
	}

	// 合并 STDOUT 和 STDERR
	output := outBuf.String()
	errOutput := errBuf.String()

	// 如果有错误输出，添加到输出中
	if errOutput != "" {
		output += "\nError output: " + errOutput
	}

	// 检查 curl 命令是否成功执行
	if strings.Contains(output, "curl: (") {
		return "", fmt.Errorf("curl command failed: %s", output)
	}

	return output, nil
}

// 执行shell命令
func (dm *DockerManager) ExecuteShellCommand(containerID, cmd string) (string, error) {
	data := url.Values{}
	data.Set("cmd", cmd)
	return dm.sendRequest(containerID, "POST", "/execute", data.Encode())
}

// 获取共识状态
func (dm *DockerManager) GetConsensusStatus(containerID string) (string, error) {
	return dm.sendRequest(containerID, "GET", "/proxy/-1/consensus", nil)
}

// 获取交易池状态
func (dm *DockerManager) GetTxpoolStatus(containerID string) (string, error) {
	return dm.sendRequest(containerID, "GET", "/proxy/-1/txpool", nil)
}

// 获取特定高度的区块
func (dm *DockerManager) GetBlockAtHeight(containerID string, height int) (string, error) {
	return dm.sendRequest(containerID, "GET", fmt.Sprintf("/proxy/-1/blocks/height/%d", height), nil)
}

// 创建本地集群工厂
func (dm *DockerManager) CreateLocalClusterFactory(containerID string, nodeCount, stakeQuota, windowSize int) (string, error) {
	body := map[string]int{
		"nodeCount":  nodeCount,
		"stakeQuota": stakeQuota,
		"windowSize": windowSize,
	}
	return dm.sendRequest(containerID, "POST", "/setup/factory", body)
}

// 创建本地点和主题地址
func (dm *DockerManager) MakeLocalAddresses(containerID string) (string, error) {
	return dm.sendRequest(containerID, "POST", "/setup/addrs", nil)
}

// 创建验证者密钥和权益配额
func (dm *DockerManager) MakeValidatorKeysAndStakeQuotas(containerID string) (string, error) {
	return dm.sendRequest(containerID, "POST", "/setup/random", nil)
}

// 写入创世文件
func (dm *DockerManager) WriteGenesisFiles(containerID string) (string, error) {
	return dm.sendRequest(containerID, "POST", "/setup/template", nil)
}

// 创建名为cluster_template的集群
func (dm *DockerManager) CreateCluster(containerID string) (string, error) {
	return dm.sendRequest(containerID, "POST", "/setup/cluster/create", nil)
}

// 构建区块链二进制文件
func (dm *DockerManager) BuildBlockchainBinary(containerID string) (string, error) {
	return dm.sendRequest(containerID, "POST", "/setup/build/chain", nil)
}

// 查看每个节点的工作目录
func (dm *DockerManager) ViewNodeWorkingDirectory(containerID string, nodeIndex int) (string, error) {
	return dm.sendRequest(containerID, "GET", fmt.Sprintf("/workdir/%d/genesis.json", nodeIndex), nil)
}

// 启动集群
func (dm *DockerManager) StartCluster(containerID string) (string, error) {
	return dm.sendRequest(containerID, "POST", "/setup/cluster/start", nil)
}

// 停止集群
func (dm *DockerManager) StopCluster(containerID string) (string, error) {
	return dm.sendRequest(containerID, "POST", "/setup/cluster/stop", nil)
}
