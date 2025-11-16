package utils

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"
)

// GetEnvironmentInfo 获取当前环境信息
func GetEnvironmentInfo() string {
	var info strings.Builder
	
	// 当前时间信息
	currentTime := time.Now()
	info.WriteString(fmt.Sprintf("当前时间: %s\n", currentTime.Format("2006-01-02 15:04:05")))
	info.WriteString(fmt.Sprintf("时区: %s\n", currentTime.Location().String()))
	
	// 操作系统信息
	info.WriteString(fmt.Sprintf("操作系统: %s\n", runtime.GOOS))
	info.WriteString(fmt.Sprintf("架构: %s\n", runtime.GOARCH))
	
	// 根据操作系统获取更详细的信息
	switch runtime.GOOS {
	case "windows":
		info.WriteString("终端类型: cmd/PowerShell\n")
		info.WriteString("命令语法: Windows命令\n")
	case "darwin":
		info.WriteString("终端类型: Terminal/zsh/bash\n")
		info.WriteString("命令语法: Unix/macOS命令\n")
	case "linux":
		info.WriteString("终端类型: bash/zsh/其他shell\n")
		info.WriteString("命令语法: Linux命令\n")
	}
	
	// 获取当前工作目录
	if wd, err := os.Getwd(); err == nil {
		info.WriteString(fmt.Sprintf("当前目录: %s\n", wd))
	}
	
	// 获取用户信息
	if user := os.Getenv("USER"); user != "" {
		info.WriteString(fmt.Sprintf("当前用户: %s\n", user))
	} else if user := os.Getenv("USERNAME"); user != "" {
		info.WriteString(fmt.Sprintf("当前用户: %s\n", user))
	}
	
	// 获取shell信息
	if shell := os.Getenv("SHELL"); shell != "" {
		info.WriteString(fmt.Sprintf("当前Shell: %s\n", shell))
	}
	
	return info.String()
}