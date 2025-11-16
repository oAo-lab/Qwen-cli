package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"time"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("用法: ask_updater <目标程序路径> <新程序路径>")
		os.Exit(1)
	}

	targetPath := os.Args[1]  // 当前运行的程序路径
	newFilePath := os.Args[2] // 下载的新程序路径

	fmt.Println("🔄 Qwen-cli 更新器启动...")
	fmt.Printf("目标程序: %s\n", targetPath)
	fmt.Printf("新程序: %s\n", newFilePath)

	// 等待主程序退出
	fmt.Println("⏳ 等待主程序退出...")
	maxWaitTime := 30 * time.Second
	startTime := time.Now()

	for {
		// 尝试重命名文件（检查文件是否还被锁定）
		err := os.Rename(newFilePath, targetPath+".new")
		if err == nil {
			// 重命名成功，说明文件已解锁
			break
		}

		// 检查是否超时
		if time.Since(startTime) > maxWaitTime {
			fmt.Printf("❌ 等待主程序退出超时，请手动关闭程序后重试\n")
			os.Exit(1)
		}

		// 等待 500ms 后重试
		time.Sleep(500 * time.Millisecond)
	}

	// 备份当前版本
	backupPath := targetPath + ".backup"
	fmt.Printf("📦 备份当前版本到: %s\n", backupPath)
	err := os.Rename(targetPath, backupPath)
	if err != nil {
		fmt.Printf("❌ 备份失败: %v\n", err)
		// 尝试恢复新文件
		os.Rename(targetPath+".new", newFilePath)
		os.Exit(1)
	}

	// 移动新版本到目标位置
	fmt.Println("🔄 安装新版本...")
	err = os.Rename(targetPath+".new", targetPath)
	if err != nil {
		fmt.Printf("❌ 安装新版本失败: %v\n", err)
		// 尝试恢复备份
		os.Rename(backupPath, targetPath)
		os.Exit(1)
	}

	// 在 Windows 上设置可执行权限
	if os.PathSeparator == '\\' {
		// Windows 不需要设置执行权限，但可以尝试
		_ = os.Chmod(targetPath, 0755)
	} else {
		// Unix 系统设置执行权限
		err = os.Chmod(targetPath, 0755)
		if err != nil {
			fmt.Printf("⚠️ 设置执行权限失败: %v\n", err)
		}
	}

	// 删除备份文件（可选）
	go func() {
		time.Sleep(5 * time.Second)
		os.Remove(backupPath)
	}()

	fmt.Println("✅ 更新完成！")

	// 启动新的程序
	fmt.Println("🚀 启动新版本...")
	cmd := exec.Command(targetPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	err = cmd.Start()
	if err != nil {
		fmt.Printf("❌ 启动新版本失败: %v\n", err)
		fmt.Printf("请手动启动: %s\n", targetPath)
		os.Exit(1)
	}

	fmt.Println("🎉 新版本已启动！")
	os.Exit(0)
}

// downloadFile 下载文件到指定路径
func downloadFile(url, filepath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("下载失败，状态码: %d", resp.StatusCode)
	}

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

// ensureDir 确保目录存在
func ensureDir(dir string) error {
	return os.MkdirAll(dir, 0755)
}
