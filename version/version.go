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
	Version   = "v0.1.0"  // 默认版本，构建时会被替换
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
			_, _ = fmt.Sscanf(newParts[i], "%d", &newNum)
		}
		if i < len(currentParts) {
			_, _ = fmt.Sscanf(currentParts[i], "%d", &currentNum)
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

	fmt.Printf("🔍 查找下载URL: 系统=%s, 架构=%s\n", osName, arch)
	fmt.Printf("🔍 总共有 %d 个资源文件\n", len(release.Assets))

	// 构建期望的文件名模式
	var expectedPattern string
	if osName == "windows" {
		expectedPattern = fmt.Sprintf("ask_.*_%s_%s.exe", osName, arch)
	} else {
		expectedPattern = fmt.Sprintf("ask_.*_%s_%s", osName, arch)
	}

	fmt.Printf("🎯 期望的文件名模式: %s\n", expectedPattern)

	// 查找匹配的资源文件
	for i, asset := range release.Assets {
		fmt.Printf("🔍 [%d] 检查文件: %s\n", i+1, asset.Name)

		// 使用更精确的匹配逻辑
		if strings.Contains(asset.Name, "ask_") &&
			strings.Contains(asset.Name, "_"+osName+"_") &&
			strings.Contains(asset.Name, "_"+arch) {

			// Windows 检查 .exe 后缀，其他平台检查无后缀或对应后缀
			if osName == "windows" {
				if strings.HasSuffix(asset.Name, ".exe") {
					fmt.Printf("✅ 找到匹配的 Windows 文件: %s\n", asset.Name)
					return asset.URL
				}
			} else {
				// macOS 和 Linux 的可执行文件通常没有后缀
				if !strings.Contains(asset.Name, ".tar.gz") && !strings.Contains(asset.Name, ".zip") {
					fmt.Printf("✅ 找到匹配的 Unix 文件: %s\n", asset.Name)
					return asset.URL
				}
			}
		}
	}

	// 如果没找到直接的可执行文件，回退到压缩包
	fmt.Printf("⚠️ 未找到直接可执行文件，尝试压缩包...\n")
	var archivePattern string
	switch osName {
	case "windows":
		archivePattern = fmt.Sprintf("_%s_%s.tar.gz", osName, arch)
	case "darwin":
		archivePattern = fmt.Sprintf("_%s_%s.tar.gz", osName, arch)
	case "linux":
		archivePattern = fmt.Sprintf("_%s_%s.tar.gz", osName, arch)
	}

	fmt.Printf("🎯 压缩包模式: %s\n", archivePattern)

	for _, asset := range release.Assets {
		if strings.Contains(asset.Name, archivePattern) {
			fmt.Printf("✅ 找到匹配的压缩包: %s\n", asset.Name)
			return asset.URL
		}
	}

	fmt.Printf("❌ 未找到任何匹配的文件\n")
	return ""
}

// DownloadAndInstall 下载并安装新版本
func DownloadAndInstall(url string) error {
	// 获取当前可执行文件路径
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("获取可执行文件路径失败: %v", err)
	}

	// 检查是否是直接的可执行文件下载（所有平台）
	isDirectBinary := strings.Contains(url, "ask_") && !strings.Contains(url, ".tar.gz")

	if isDirectBinary {
		// 所有平台都直接下载可执行文件
		return downloadAndInstallBinary(url, execPath)
	} else {
		// 下载压缩包并安装
		return downloadAndInstallArchive(url, execPath)
	}
}

// downloadAndInstallBinary 下载并安装直接的可执行文件（用于所有平台）
func downloadAndInstallBinary(url, execPath string) error {
	fmt.Println("📦 正在下载可执行文件...")

	// 根据操作系统确定临时文件扩展名
	var ext string
	if runtime.GOOS == "windows" {
		ext = ".exe"
	} else {
		ext = ""
	}

	// 创建临时文件
	tmpFile, err := os.CreateTemp("", "qwen-cli-update-*"+ext)
	if err != nil {
		return fmt.Errorf("创建临时文件失败: %v", err)
	}
	defer func() {
		_ = os.Remove(tmpFile.Name())
	}()
	defer func() {
		_ = tmpFile.Close()
	}()

	// 下载文件
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("下载失败: %v", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("下载失败，状态码: %d", resp.StatusCode)
	}

	// 写入临时文件
	_, err = io.Copy(tmpFile, resp.Body)
	if err != nil {
		return fmt.Errorf("写入文件失败: %v", err)
	}

	if runtime.GOOS == "windows" {
		// 在Windows上，使用外部更新器程序
		return downloadAndInstallWithUpdater(url, execPath, tmpFile.Name())
	} else {
		// 在Unix系统上，直接替换可执行文件
		// 备份当前版本
		backupPath := execPath + ".backup"
		err = os.Rename(execPath, backupPath)
		if err != nil {
			return fmt.Errorf("备份当前版本失败: %v", err)
		}

		// 移动新版本到目标位置
		err = os.Rename(tmpFile.Name(), execPath)
		if err != nil {
			// 如果失败，恢复备份
			_ = os.Rename(backupPath, execPath)
			return fmt.Errorf("替换文件失败: %v", err)
		}

		// 设置执行权限
		err = os.Chmod(execPath, 0755)
		if err != nil {
			return fmt.Errorf("设置执行权限失败: %v", err)
		}

		// 删除备份文件
		_ = os.Remove(backupPath)

		fmt.Println("✅ 更新完成！")
		fmt.Println("🔄 请重新启动 Qwen-cli 以使用新版本")
	}

	return nil
}

// downloadAndInstallArchive 下载压缩包并安装（备用方案）
func downloadAndInstallArchive(url, execPath string) error {
	fmt.Println("📦 正在下载压缩包...")

	// 创建临时文件
	tmpFile, err := os.CreateTemp("", "qwen-cli-update-*.tar.gz")
	if err != nil {
		return fmt.Errorf("创建临时文件失败: %v", err)
	}
	defer func() {
		_ = os.Remove(tmpFile.Name())
	}()
	defer func() {
		_ = tmpFile.Close()
	}()

	// 下载文件
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("下载失败: %v", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("下载失败，状态码: %d", resp.StatusCode)
	}

	// 写入临时文件
	_, err = io.Copy(tmpFile, resp.Body)
	if err != nil {
		return fmt.Errorf("写入文件失败: %v", err)
	}

	// 在Windows上，如果下载的是压缩包，也进行自动处理
	if runtime.GOOS == "windows" {
		// 解压到临时目录
		fmt.Println("📦 正在解压更新包...")

		// 创建临时目录
		tmpDir, err := os.MkdirTemp("", "qwen-cli-update-*")
		if err != nil {
			return fmt.Errorf("创建临时目录失败: %v", err)
		}
		defer func() {
			_ = os.RemoveAll(tmpDir)
		}()

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
			if !file.IsDir() && strings.HasSuffix(file.Name(), ".exe") {
				binaryPath = filepath.Join(tmpDir, file.Name())
				break
			}
		}

		if binaryPath == "" {
			return fmt.Errorf("在更新包中找不到可执行文件")
		}

		// 使用外部更新器程序
		_ = downloadAndInstallWithUpdater("", execPath, binaryPath)
		return nil
	}

	// 在Unix系统上，自动解压并替换文件
	fmt.Println("📦 正在解压更新包...")

	// 创建临时目录
	tmpDir, err := os.MkdirTemp("", "qwen-cli-update-*")
	if err != nil {
		return fmt.Errorf("创建临时目录失败: %v", err)
	}
	defer func() {
		_ = os.RemoveAll(tmpDir)
	}()

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
		_ = os.Rename(backupPath, execPath)
		return fmt.Errorf("替换文件失败: %v", err)
	}

	// 设置执行权限
	err = os.Chmod(execPath, 0755)
	if err != nil {
		return fmt.Errorf("设置执行权限失败: %v", err)
	}

	// 删除备份文件
	_ = os.Remove(backupPath)

	fmt.Println("✅ 更新完成！")
	fmt.Println("🔄 请重新启动 Qwen-cli 以使用新版本")

	return nil
}

// downloadAndInstallWithUpdater 使用外部更新器程序进行更新（Windows专用）
func downloadAndInstallWithUpdater(url, execPath, newBinaryPath string) error {
	fmt.Println("🔄 准备使用外部更新器...")

	// 如果提供了URL，需要先下载
	if url != "" {
		fmt.Println("📦 正在下载更新文件...")
		resp, err := http.Get(url)
		if err != nil {
			return fmt.Errorf("下载失败: %v", err)
		}
		defer func() {
			_ = resp.Body.Close()
		}()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("下载失败，状态码: %d", resp.StatusCode)
		}

		// 创建临时文件保存下载的内容
		tmpFile, err := os.CreateTemp("", "qwen-cli-update-*.exe")
		if err != nil {
			return fmt.Errorf("创建临时文件失败: %v", err)
		}
		defer func() {
			_ = tmpFile.Close()
		}()

		// 写入下载内容
		_, err = io.Copy(tmpFile, resp.Body)
		if err != nil {
			return fmt.Errorf("写入文件失败: %v", err)
		}
		newBinaryPath = tmpFile.Name()
	}

	// 获取或下载更新器程序
	updaterPath, err := getOrUpdateUpdater()
	if err != nil {
		return fmt.Errorf("获取更新器失败: %v", err)
	}

	fmt.Printf("🚀 启动更新器: %s\n", updaterPath)

	// 启动更新器程序
	cmd := exec.Command(updaterPath, execPath, newBinaryPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Start()
	if err != nil {
		return fmt.Errorf("启动更新器失败: %v", err)
	}

	fmt.Println("✅ 更新程序已启动，将在几秒钟内完成...")
	fmt.Println("🔄 请等待更新完成后重新启动 Qwen-cli")

	// 立即退出当前程序，释放文件锁定
	os.Exit(0)
	return nil
}

// getOrUpdateUpdater 获取或下载更新器程序
func getOrUpdateUpdater() (string, error) {
	// 首先尝试在当前目录查找
	execDir := filepath.Dir(getExecutablePath())
	updaterPath := filepath.Join(execDir, "ask_updater.exe")
	if _, err := os.Stat(updaterPath); err == nil {
		return updaterPath, nil
	}

	// 尝试在临时目录查找
	tempDir := os.TempDir()
	updaterPath = filepath.Join(tempDir, "ask_updater.exe")
	if _, err := os.Stat(updaterPath); err == nil {
		return updaterPath, nil
	}

	// 如果都找不到，下载更新器
	fmt.Println("📦 正在下载更新器程序...")
	return downloadUpdater()
}

// downloadUpdater 下载更新器程序
func downloadUpdater() (string, error) {
	// 获取当前版本信息，用于下载对应版本的更新器
	currentVersion := Version
	if currentVersion == "v0.1.0" || currentVersion == "unknown" {
		// 如果是默认版本，尝试获取最新版本
		if release, err := getLatestRelease(); err == nil {
			currentVersion = release.TagName
		}
	}

	// 构建更新器下载URL，根据实际发布结构调整
	// 从GitHub Releases可以看到文件名格式为：ask_updater_0.1.22_windows_amd64.exe
	versionWithoutV := strings.TrimPrefix(currentVersion, "v")
	updaterURL := fmt.Sprintf("https://github.com/oAo-lab/Qwen-cli/releases/download/%s/ask_updater_%s_windows_%s.exe",
		currentVersion, versionWithoutV, runtime.GOARCH)

	// 创建临时文件保存更新器
	tempDir := os.TempDir()
	updaterPath := filepath.Join(tempDir, "ask_updater.exe")

	fmt.Printf("📥 正在下载更新器: %s\n", updaterURL)

	// 下载更新器
	resp, err := http.Get(updaterURL)
	if err != nil {
		return "", fmt.Errorf("下载更新器失败: %v", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("下载更新器失败，状态码: %d，URL: %s", resp.StatusCode, updaterURL)
	}

	// 保存更新器文件
	out, err := os.Create(updaterPath)
	if err != nil {
		return "", fmt.Errorf("创建更新器文件失败: %v", err)
	}
	defer func() {
		_ = out.Close()
	}()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return "", fmt.Errorf("保存更新器失败: %v", err)
	}

	// 设置执行权限
	err = os.Chmod(updaterPath, 0755)
	if err != nil {
		return "", fmt.Errorf("设置更新器权限失败: %v", err)
	}

	fmt.Println("✅ 更新器下载完成")
	return updaterPath, nil
}

// getLatestRelease 获取最新发布信息
func getLatestRelease() (*ReleaseInfo, error) {
	url := "https://api.github.com/repos/oAo-lab/Qwen-cli/releases/latest"
	fmt.Printf("🔍 正在请求: %s\n", url)

	// 添加重试机制处理速率限制
	maxRetries := 3
	retryDelay := time.Second * 2

	for attempt := 1; attempt <= maxRetries; attempt++ {
		fmt.Printf("🔄 尝试第 %d 次...\n", attempt)

		// 创建 HTTP 客户端，设置 User-Agent
		client := &http.Client{
			Timeout: 30 * time.Second,
		}

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			fmt.Printf("❌ 创建请求失败: %v\n", err)
			return nil, err
		}

		// 设置 User-Agent 避免被限制
		req.Header.Set("User-Agent", "Qwen-cli/"+Version)
		req.Header.Set("Accept", "application/vnd.github.v3+json")

		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("❌ HTTP 请求失败: %v\n", err)
			if attempt < maxRetries {
				fmt.Printf("⏳ %d 秒后重试...\n", retryDelay)
				time.Sleep(retryDelay)
				retryDelay *= 2 // 指数退避
				continue
			}
			return nil, err
		}
		defer func() {
			_ = resp.Body.Close()
		}()

		fmt.Printf("📊 响应状态码: %d\n", resp.StatusCode)

		// 处理速率限制
		if resp.StatusCode == 403 {
			fmt.Printf("⚠️ 可能遇到 GitHub API 速率限制\n")
			if attempt < maxRetries {
				resetTime := resp.Header.Get("X-RateLimit-Reset")
				if resetTime != "" {
					// 解析 Unix 时间戳
					if timestamp, err := time.Parse("2006-01-02T15:04:05Z", resetTime); err == nil {
						waitTime := time.Until(timestamp)
						if waitTime > 0 {
							fmt.Printf("📅 速率限制将在 %s 后重置\n", timestamp.Format("15:04:05"))
							fmt.Printf("⏳ 需要等待 %.0f 分钟\n", waitTime.Minutes())
							retryDelay = waitTime
						}
					} else {
						fmt.Printf("📅 速率限制重置时间: %s\n", resetTime)
					}
				}
				fmt.Printf("⏳ %d 秒后重试...\n", retryDelay)
				time.Sleep(retryDelay)
				retryDelay *= 2
				continue
			}
			return nil, fmt.Errorf("GitHub API 速率限制，请稍后再试")
		}

		// 处理其他错误状态码
		if resp.StatusCode != 200 {
			if attempt < maxRetries {
				fmt.Printf("⏳ %d 秒后重试...\n", retryDelay)
				time.Sleep(retryDelay)
				retryDelay *= 2
				continue
			}
			return nil, fmt.Errorf("API 请求失败，状态码: %d", resp.StatusCode)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("❌ 读取响应失败: %v\n", err)
			if attempt < maxRetries {
				fmt.Printf("⏳ %d 秒后重试...\n", retryDelay)
				time.Sleep(retryDelay)
				retryDelay *= 2
				continue
			}
			return nil, err
		}

		fmt.Printf("📄 响应内容长度: %d 字节\n", len(body))

		var release ReleaseInfo
		if err := json.Unmarshal(body, &release); err != nil {
			fmt.Printf("❌ JSON 解析失败: %v\n", err)
			fmt.Printf("📝 原始响应: %s\n", string(body))
			return nil, err
		}

		fmt.Printf("✅ 成功解析发布信息: %s\n", release.TagName)
		return &release, nil
	}

	return nil, fmt.Errorf("所有重试均失败")
}

// GetLatestRelease 获取最新发布信息（导出版本）
func GetLatestRelease() (*ReleaseInfo, error) {
	return getLatestRelease()
}

// getExecutablePath 获取当前可执行文件路径
func getExecutablePath() string {
	execPath, err := os.Executable()
	if err != nil {
		return ""
	}
	return execPath
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
