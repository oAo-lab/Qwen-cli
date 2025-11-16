package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"Qwen-cli/commands"
	"Qwen-cli/config"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "ask",
		Short: "通义千问命令行客户端",
		Long:  `通义千问命令行客户端，支持多模型对话和角色切换。`,
	}

	// 添加 init 命令（不需要配置）
	rootCmd.AddCommand(commands.InitCommand())

	// 尝试加载配置并添加需要配置的命令
	cfg, err := config.LoadConfig()
	if err != nil {
		// 如果配置加载失败，只显示提示信息
		fmt.Printf("⚠️  配置文件未找到或加载失败: %s\n", err)
		fmt.Println("💡 请运行 'ask init' 初始化配置文件")
		fmt.Println()
	} else {
		// 配置加载成功，添加需要配置的命令
		rootCmd.AddCommand(commands.ChatCommand(cfg))
		rootCmd.AddCommand(commands.CmdCommand(cfg))
		rootCmd.AddCommand(commands.TestCommand(cfg))
		rootCmd.AddCommand(commands.DebugCommand(cfg))
	}

	// Handle SIGINT signal to pause the conversation
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT)

	go func() {
		for range signalChan {
			fmt.Println("\n对话已暂停，按回车键继续...")
			fmt.Scanln()
		}
	}()

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
