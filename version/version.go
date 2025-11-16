package version

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
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

	var pattern string
	switch osName {
	case "windows":
		if arch == "arm64" {
			// Windows 优先尝试直接下载 exe 文件
			pattern = "ask_"
		} else {
			// Windows 优先尝试直接下载 exe 文件
			pattern = "ask_"
		}
	case "darwin":
		if arch == "arm64" {
			pattern = "_darwin_arm64.tar.gz"
		} else {
			pattern = "_darwin_amd64.tar.gz"
		}
	case "linux":
		if arch == "arm64" {
			pattern = "_linux_arm64.tar.gz"
		} else {
			pattern = "_linux_amd64.tar.gz"
		}
	default:
		return ""
	}

	// 查找匹配的资源文件
	for _, asset := range release.Assets {
		if osName == "windows" {
			// Windows 特殊处理：优先查找直接的可执行文件
			if strings.Contains(asset.Name, "ask_") && strings.Contains(asset.Name, "_windows_") &&
			   (strings.HasSuffix(asset.Name, ".exe") || !strings.Contains(asset.Name, ".")) {
				return asset.URL
			}
		} else {
			// 其他系统使用原有的压缩包逻辑
			if strings.Contains(asset.Name, pattern) {
				return asset.URL
			}
		}
	}

	// 如果 Windows 没找到直接的可执行文件，回退到压缩包
	if osName == "windows" {
		if arch == "arm64" {
			pattern = "_windows_arm64.tar.gz"
		} else {
			pattern = "_windows_amd64.tar.gz"
		}
		for _, asset := range release.Assets {
			if strings.Contains(asset.Name, pattern) {
				return asset.URL
			}
		}
	}

	return ""
}

// DownloadAndInstall 下载并安装新版本
func DownloadAndInstall(url string) error {
	// 获取当前可执行文件路径
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("获取可执行文件路径失败: %v", err)
	}

	// 检查是否是直接的可执行文件下载（Windows）
	isDirectBinary := runtime.GOOS == "windows" &&
		(strings.Contains(url, "ask_") && !strings.Contains(url, ".tar.gz"))

	if isDirectBinary {
		// Windows 直接下载可执行文件
		return downloadAndInstallBinary(url, execPath)
	} else {
		// 下载压缩包并安装
		return downloadAndInstallArchive(url, execPath)
	}
}

// downloadAndInstallBinary 下载并安装直接的可执行文件（主要用于Windows）
func downloadAndInstallBinary(url, execPath string) error {
	fmt.Println("📦 正在下载可执行文件...")

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

	// 在Windows上，需要创建一个批处理文件来替换可执行文件
	batchContent := fmt.Sprintf(`@echo off
echo 正在更新 Qwen-cli...
timeout /t 2 /nobreak >nul
move /Y "%s" "%s" >nul 2>&1
echo 更新完成！
del "%%~f0"
`, tmpFile.Name(), execPath)

	// 创建批处理文件
	batchFile := execPath + ".update.bat"
	err = os.WriteFile(batchFile, []byte(batchContent), 0644)
	if err != nil {
		return fmt.Errorf("创建更新脚本失败: %v", err)
	}

	// 启动批处理文件并退出当前程序
	cmd := exec.Command("cmd", "/C", batchFile)
	cmd.Start()

	fmt.Println("✅ 更新程序已启动，将在几秒钟内完成...")
	fmt.Println("🔄 请重新启动 Qwen-cli 以使用新版本")

	return nil
}

// downloadAndInstallArchive 下载压缩包并安装（用于Unix系统和Windows压缩包）
func downloadAndInstallArchive(url, execPath string) error {
	// 创建临时文件
	tmpFile, err := os.CreateTemp("", "qwen-cli-update-*.tar.gz")
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

	// 在Windows上，如果下载的是压缩包，仍然需要手动处理
	if runtime.GOOS == "windows" {
		fmt.Println("在Windows上更新需要手动替换文件...")
		fmt.Printf("请手动下载并解压以下文件: %s\n", url)
		fmt.Printf("然后将解压后的可执行文件替换当前程序: %s\n", execPath)
		return nil
	}

	// 在Unix系统上，自动解压并替换文件
	fmt.Println("📦 正在解压更新包...")
	
	// 创建临时目录
	tmpDir, err := os.MkdirTemp("", "qwen-cli-update-*")
	if err != nil {
		return fmt.Errorf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// 解压文件
	err = extractTarGz(tmpFile.Name(), tmpDir)
	if err != nil {
		return fmt.Errorf("解压失败: %v", err)
	}

	// 查找解压后的可执行文件
	var binaryPath string
	files, err := os.ReadDir(tmpDir)
	if err != nil {
		return fmt.Errorf("读取解压目录失败: %v", err)
	}

	for _, file := range files {
		if !file.IsDir() && (file.Name() == "ask" || (runtime.GOOS == "windows" && file.Name() == "ask.exe")) {
			binaryPath = tmpDir + "/" + file.Name()
			break
		}
	}

	if binaryPath == "" {
		return fmt.Errorf("在更新包中找不到可执行文件")
	}

	// 备份当前版本
	backupPath := execPath + ".backup"
	err = os.Rename(execPath, backupPath)
	if err != nil {
		return fmt.Errorf("备份当前版本失败: %v", err)
	}

	// 移动新版本到目标位置
	err = os.Rename(binaryPath, execPath)
	if err != nil {
		// 如果失败，恢复备份
		os.Rename(backupPath, execPath)
		return fmt.Errorf("替换文件失败: %v", err)
	}

	// 设置执行权限
	err = os.Chmod(execPath, 0755)
	if err != nil {
		return fmt.Errorf("设置执行权限失败: %v", err)
	}

	// 删除备份文件
	os.Remove(backupPath)

	return nil
}

// extractTarGz 解压 tar.gz 文件到指定目录
func extractTarGz(src, dest string) error {
	// 打开 gzip 文件
	gzFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("打开gzip文件失败: %v", err)
	}
	defer gzFile.Close()

	// 创建 gzip reader
	gzReader, err := gzip.NewReader(gzFile)
	if err != nil {
		return fmt.Errorf("创建gzip reader失败: %v", err)
	}
	defer gzReader.Close()

	// 创建 tar reader
	tarReader := tar.NewReader(gzReader)

	// 遍历 tar 文件中的每个文件
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break // 文件结束
		}
		if err != nil {
			return fmt.Errorf("读取tar文件失败: %v", err)
		}

		// 构建目标文件路径
		targetPath := filepath.Join(dest, header.Name)

		// 根据文件类型进行处理
		switch header.Typeflag {
		case tar.TypeDir:
			// 创建目录
			if err := os.MkdirAll(targetPath, os.FileMode(header.Mode)); err != nil {
				return fmt.Errorf("创建目录失败: %v", err)
			}
		case tar.TypeReg:
			// 创建文件
			outFile, err := os.OpenFile(targetPath, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return fmt.Errorf("创建文件失败: %v", err)
			}

			// 复制文件内容
			if _, err := io.Copy(outFile, tarReader); err != nil {
				outFile.Close()
				return fmt.Errorf("写入文件失败: %v", err)
			}
			outFile.Close()
		}
	}

	return nil
}