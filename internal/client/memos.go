package client

import (
	"crypto/tls" // 新增：用于TLS配置
	"encoding/json"
	"fmt"
	"os"

	"exporter-to-obsidian/internal/types"

	"github.com/go-resty/resty/v2"
	"github.com/joho/godotenv"
)

// MemosClient Memos API客户端
type MemosClient struct {
	apiURL string
	token  string
	client *resty.Client
}

// NewMemosClient 创建新的Memos客户端
func NewMemosClient(apiURL, token string) (*MemosClient, error) {
	// 加载.env文件
	godotenv.Load()

	if apiURL == "" {
		apiURL = os.Getenv("MEMOS_API")
	}
	if token == "" {
		token = os.Getenv("MEMOS_TOKEN")
	}

	if apiURL == "" || token == "" {
		return nil, fmt.Errorf("请提供Memos API URL和Token")
	}

	client := &MemosClient{
		apiURL: apiURL,
		token:  token,
		client: resty.New().SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true}), // 修改：添加跳过TLS验证
	}

	// 设置默认请求头
	client.client.SetHeaders(map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", token),
		"Accept":        "application/json",
	})

	return client, nil
}

// FetchMemos 获取Memos数据
func (c *MemosClient) FetchMemos(limit, offset int, rowStatus string) ([]types.MemosRecord, error) {
	resp, err := c.client.R().
		SetQueryParams(map[string]string{
			"limit":     fmt.Sprintf("%d", limit),
			"offset":    fmt.Sprintf("%d", offset),
			"rowStatus": rowStatus,
		}).
		Get(c.apiURL)

	if err != nil {
		return nil, fmt.Errorf("获取Memos数据失败: %v", err)
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("获取Memos数据失败，状态码: %d", resp.StatusCode())
	}

	var records []types.MemosRecord
	if err := json.Unmarshal(resp.Body(), &records); err != nil {
		return nil, fmt.Errorf("解析Memos数据失败: %v", err)
	}

	return records, nil
}
