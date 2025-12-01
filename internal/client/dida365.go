package client

import (
	"crypto/tls" // 新增：用于TLS配置
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time" // 新增：用于时间处理

	"exporter-to-obsidian/internal/types"

	"github.com/go-resty/resty/v2"
	"github.com/joho/godotenv"
)

// Dida365Client 滴答清单API客户端
type Dida365Client struct {
	username      string
	password      string
	baseURL       string
	client        *resty.Client
	token         string
	inboxID       string
	lastLoginTime time.Time // 新增：存储上次登录时间
}

// NewDida365Client 创建新的滴答清单客户端
func NewDida365Client(username, password string) (*Dida365Client, error) {
	// 加载.env文件
	godotenv.Load()

	if username == "" {
		username = os.Getenv("DIDA365_USERNAME")
	}
	if password == "" {
		password = os.Getenv("DIDA365_PASSWORD")
	}

	if username == "" || password == "" {
		return nil, fmt.Errorf("请提供账号信息。可以通过参数传入或设置环境变量：\nDIDA365_USERNAME: 你的滴答清单用户名/邮箱\nDIDA365_PASSWORD: 你的滴答清单密码")
	}

	client := &Dida365Client{
		username: username,
		password: password,
		baseURL:  "https://api.dida365.com/api/v2",
		client:   resty.New().SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true}), // 修改：添加跳过TLS验证
	}

	// 设置默认请求头
	client.client.SetHeaders(map[string]string{
		"user-agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36",
		"x-device":   `{"platform":"web","os":"Windows 11","device":"Chrome 131.0.0.0","name":"","version":6246,"id":"674ea3c2a4f37a3f2c9b42d8","channel":"website","campaign":"","websocket":"67e7de9bf92b296c741567e0"}`,
	})

	// 尝试从环境变量加载token和上次登录时间
	client.loadTokenFromEnv()
	if client.token == "" {
		if err := client.Login(); err != nil {
			return nil, err
		}
	} else {
		// 检查上次登录时间是否超过一天
		if time.Since(client.lastLoginTime) > 24*time.Hour {
			fmt.Println("Token已过期（超过一天），重新登录")
			if err := client.Login(); err != nil {
				return nil, err
			}
		} else {
			fmt.Println("使用本地保存的Token")
			client.client.SetHeader("Cookie", fmt.Sprintf("t=%s", client.token))
		}
	}

	return client, nil
}

// loadTokenFromEnv 从环境变量加载token和上次登录时间
func (c *Dida365Client) loadTokenFromEnv() {
	c.token = os.Getenv("DIDA365_TOKEN")
	c.inboxID = os.Getenv("DIDA365_INBOX_ID")
	if c.token == "None" {
		c.token = ""
	}
	if c.inboxID == "None" {
		c.inboxID = ""
	}

	// 加载上次登录时间
	lastLoginStr := os.Getenv("DIDA365_LAST_LOGIN_TIME")
	if lastLoginStr != "" {
		t, err := time.Parse(time.RFC3339, lastLoginStr)
		if err == nil {
			c.lastLoginTime = t
		}
	}
}

// saveTokenToEnv 保存token、上次登录时间到环境变量文件
func (c *Dida365Client) saveTokenToEnv() error {
	envFile := ".env"
	if _, err := os.Stat(envFile); os.IsNotExist(err) {
		// 如果.env文件不存在，创建一个
		file, err := os.Create(envFile)
		if err != nil {
			return fmt.Errorf("创建.env文件失败: %v", err)
		}
		file.Close()
	}

	// 读取现有内容
	content, err := os.ReadFile(envFile)
	if err != nil {
		return fmt.Errorf("读取.env文件失败: %v", err)
	}

	lines := strings.Split(string(content), "\n")
	updatedToken := false
	updatedInbox := false
	updatedTime := false

	// 更新或添加token、上次登录时间
	for i, line := range lines {
		if strings.HasPrefix(line, "DIDA365_TOKEN=") {
			lines[i] = fmt.Sprintf("DIDA365_TOKEN=%s", c.token)
			updatedToken = true
		} else if strings.HasPrefix(line, "DIDA365_INBOX_ID=") {
			lines[i] = fmt.Sprintf("DIDA365_INBOX_ID=%s", c.inboxID)
			updatedInbox = true
		} else if strings.HasPrefix(line, "DIDA365_LAST_LOGIN_TIME=") {
			lines[i] = fmt.Sprintf("DIDA365_LAST_LOGIN_TIME=%s", time.Now().Format(time.RFC3339))
			updatedTime = true
		}
	}

	if !updatedToken {
		lines = append(lines, fmt.Sprintf("DIDA365_TOKEN=%s", c.token))
	}
	if !updatedInbox {
		lines = append(lines, fmt.Sprintf("DIDA365_INBOX_ID=%s", c.inboxID))
	}
	if !updatedTime {
		lines = append(lines, fmt.Sprintf("DIDA365_LAST_LOGIN_TIME=%s", time.Now().Format(time.RFC3339)))
	}

	return os.WriteFile(envFile, []byte(strings.Join(lines, "\n")), 0644)
}

// Login 登录获取token并更新登录时间
func (c *Dida365Client) Login() error {
	fmt.Println("登录获取Token")
	url := fmt.Sprintf("%s/user/signon?wc=true&remember=true", c.baseURL)
	payload := map[string]string{
		"password": c.password,
		"username": c.username,
	}

	resp, err := c.client.R().
		SetBody(payload).
		Post(url)

	if err != nil {
		return fmt.Errorf("登录请求失败: %v", err)
	}

	if resp.StatusCode() != 200 {
		return fmt.Errorf("登录失败，状态码: %d, 响应: %s", resp.StatusCode(), resp.String())
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return fmt.Errorf("解析登录响应失败: %v", err)
	}

	token, ok := result["token"].(string)
	if !ok {
		return fmt.Errorf("登录响应中未找到token")
	}

	inboxID, ok := result["inboxId"].(string)
	if !ok {
		return fmt.Errorf("登录响应中未找到inboxId")
	}

	c.token = token
	c.inboxID = inboxID
	c.client.SetHeader("Cookie", fmt.Sprintf("t=%s", c.token))
	c.lastLoginTime = time.Now() // 更新登录时间

	// 保存token和登录时间到环境变量文件
	if err := c.saveTokenToEnv(); err != nil {
		fmt.Printf("保存token到.env失败: %v\n", err)
	}

	return nil
}

// GetProjects 获取所有项目列表
func (c *Dida365Client) GetProjects() ([]types.Project, error) {
	resp, err := c.client.R().
		Get(fmt.Sprintf("%s/projects", c.baseURL))

	if err != nil {
		return nil, fmt.Errorf("获取项目列表失败: %v", err)
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("获取项目列表失败，状态码: %d", resp.StatusCode())
	}

	var projects []types.Project
	if err := json.Unmarshal(resp.Body(), &projects); err != nil {
		return nil, fmt.Errorf("解析项目列表失败: %v", err)
	}

	return projects, nil
}

// GetAllData 获取项目列表、任务列表、标签列表
func (c *Dida365Client) GetAllData() (map[string]interface{}, error) {
	resp, err := c.client.R().
		Get(fmt.Sprintf("%s/batch/check/0", c.baseURL))

	if err != nil {
		return nil, fmt.Errorf("获取所有数据失败: %v", err)
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("获取所有数据失败，状态码: %d", resp.StatusCode())
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, fmt.Errorf("解析所有数据失败: %v", err)
	}

	return result, nil
}

// GetCompletedTasks 获取已完成任务列表
func (c *Dida365Client) GetCompletedTasks(fromDate, toDate string, limit int) ([]types.Task, error) {
	resp, err := c.client.R().
		SetQueryParams(map[string]string{
			"from":  fromDate,
			"to":    toDate,
			"limit": fmt.Sprintf("%d", limit),
		}).
		Get(fmt.Sprintf("%s/project/all/completed", c.baseURL))

	if err != nil {
		return nil, fmt.Errorf("获取已完成任务失败: %v", err)
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("获取已完成任务失败，状态码: %d", resp.StatusCode())
	}

	var tasks []types.Task
	if err := json.Unmarshal(resp.Body(), &tasks); err != nil {
		return nil, fmt.Errorf("解析已完成任务失败: %v", err)
	}

	return tasks, nil
}

// GetHabits 获取习惯列表
func (c *Dida365Client) GetHabits() ([]types.Habit, error) {
	resp, err := c.client.R().
		Get(fmt.Sprintf("%s/habits", c.baseURL))

	if err != nil {
		return nil, fmt.Errorf("获取习惯列表失败: %v", err)
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("获取习惯列表失败，状态码: %d", resp.StatusCode())
	}

	var habits []types.Habit
	if err := json.Unmarshal(resp.Body(), &habits); err != nil {
		return nil, fmt.Errorf("解析习惯列表失败: %v", err)
	}

	return habits, nil
}

// GetHabitsCheckins 获取习惯打卡列表
func (c *Dida365Client) GetHabitsCheckins(afterStamp string, habitIDs []string) (*types.HabitCheckinsResponse, error) {
	payload := map[string]interface{}{
		"afterStamp": afterStamp,
		"habitIds":   habitIDs,
	}

	resp, err := c.client.R().
		SetBody(payload).
		Post(fmt.Sprintf("%s/habitCheckins/query", c.baseURL))

	if err != nil {
		return nil, fmt.Errorf("获取习惯打卡失败: %v", err)
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("获取习惯打卡失败，状态码: %d", resp.StatusCode())
	}

	var result types.HabitCheckinsResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, fmt.Errorf("解析习惯打卡失败: %v", err)
	}

	return &result, nil
}

// GetInboxID 获取收集箱ID
func (c *Dida365Client) GetInboxID() string {
	return c.inboxID
}

// GetProjectColumns 获取项目分组信息
func (c *Dida365Client) GetProjectColumns(projectID string) ([]types.Column, error) {
	url := fmt.Sprintf("%s/column/project/%s", c.baseURL, projectID)
	
	resp, err := c.client.R().
		Get(url)

	if err != nil {
		return nil, fmt.Errorf("获取项目列信息失败: %v", err)
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("获取项目列信息失败，状态码: %d", resp.StatusCode())
	}

	var columns []types.Column
	if err := json.Unmarshal(resp.Body(), &columns); err != nil {
		return nil, fmt.Errorf("解析项目列信息失败: %v", err)
	}

	return columns, nil
}
