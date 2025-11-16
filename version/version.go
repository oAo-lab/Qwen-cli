package version

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"
)

// 版本信息
var (
	Version   = "v0.1.0" // 默认版本，构建时会被替换
	BuildDate = "unknown" // 构建日期，构建时会被替换
	GitCommit = "unknown" // Git提交哈希，构建时会被替换
)

// ReleaseInfo 表示GitHub发布信息
type ReleaseInfo struct {
	TagName string `json:"tag_name"`
	Name    string `json:"name"`
	Body    string `json:"body"`
	Assets  []struct {
		Name string `json:"name"`
		URL  string `json:"browser_download_url"`
	} `json:"assets"`
	PublishedAt time.Time `json:"published_at"`
}

// GetVersion 返回当前版本信息
func GetVersion() string {
	return Version
}

// GetVersionInfo 返回详细的版本信息
func GetVersionInfo() string {
	return fmt.Sprintf("Qwen-cli %s\n构建时间: %s\nGit提交: %s\nGo版本: %s\n系统: %s/%s", 
		Version, BuildDate, GitCommit, runtime.Version(), runtime.GOOS, runtime.GOARCH)
}

// CheckUpdate 检查是否有新版本
func CheckUpdate() (bool, *ReleaseInfo, error) {
	// 获取最新发布信息
	resp, err := http.Get("https://api.github.com/repos/oAo-lab/Qwen-cli/releases/latest")
	if err != nil {
		return false, nil, fmt.Errorf("获取发布信息失败: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, nil, fmt.Errorf("读取响应失败: %v", err)
	}

	var release ReleaseInfo
	if err := json.Unmarshal(body, &release); err != nil {
		return false, nil, fmt.Errorf("解析发布信息失败: %v", err)
	}

	// 比较版本号
	if isNewerVersion(release.TagName, Version) {
		return true, &release, nil
	}

	return false, &release, nil
}

// isNewerVersion 检查新版本是否比当前版本新
func isNewerVersion(newVersion, currentVersion string) bool {
	// 移除版本号前的 'v' 前缀
	newVersion = strings.TrimPrefix(newVersion, "v")
	currentVersion = strings.TrimPrefix(currentVersion, "v")

	newParts := strings.Split(newVersion, ".")
	currentParts := strings.Split(currentVersion, ".")

	// 确保版本号长度一致
	maxLen := len(newParts)
	if len(currentParts) > maxLen {
		maxLen = len(currentParts)
	}

	for i := 0; i < maxLen; i++ {
		var newNum, currentNum int

		if i < len(newParts) {
			fmt.Sscanf(newParts[i], "%d", &newNum)
		}
		if i < len(currentParts) {
			fmt.Sscanf(currentParts[i], "%d", &currentNum)
		}

		if newNum > currentNum {
			return true
		} else if newNum < currentNum {
			return false
		}
	}

	return false
}

// GetDownloadURL 根据当前系统获取合适的下载URL
func GetDownloadURL(release *ReleaseInfo) string {
	osName := runtime.GOOS
	arch := runtime.GOARCH

	var suffix string
	switch osName {
	case "windows":
		suffix = "windows-amd64.exe"
	case "darwin":
		if arch == "arm64" {
			suffix = "darwin-arm64"
		} else {
			suffix = "darwin-amd64"
		}
	case "linux":
		if arch == "arm64" {
			suffix = "linux-arm64"
		} else {
			suffix = "linux-amd64"
		}
	default:
		return ""
	}

	// 查找匹配的资源文件
	for _, asset := range release.Assets {
		if strings.HasSuffix(asset.Name, suffix) {
			return asset.URL
		}
	}

	return ""
}

// DownloadAndInstall 下载并安装新版本
func DownloadAndInstall(url string) error {
	// 创建临时文件
	tmpFile, err := os.CreateTemp("", "qwen-cli-update-*.exe")
	if err != nil {
		return fmt.Errorf("创建临时文件失败: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// 下载文件
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("下载失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("下载失败，状态码: %d", resp.StatusCode)
	}

	// 写入临时文件
	_, err = io.Copy(tmpFile, resp.Body)
	if err != nil {
		return fmt.Errorf("写入文件失败: %v", err)
	}

	// 获取当前可执行文件路径
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("获取可执行文件路径失败: %v", err)
	}

	// 在Windows上，需要先关闭当前程序才能替换文件
	if runtime.GOOS == "windows" {
		fmt.Println("在Windows上更新需要手动替换文件...")
		fmt.Printf("请将以下文件替换当前程序: %s\n", tmpFile.Name())
		fmt.Printf("当前程序位置: %s\n", execPath)
		return nil
	}

	// 在Unix系统上，可以直接替换文件
	err = os.Rename(tmpFile.Name(), execPath)
	if err != nil {
		return fmt.Errorf("替换文件失败: %v", err)
	}

	// 设置执行权限
	err = os.Chmod(execPath, 0755)
	if err != nil {
		return fmt.Errorf("设置执行权限失败: %v", err)
	}

	return nil
}