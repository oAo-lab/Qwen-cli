package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"Qwen-cli/config"
)

func InitCommand() *cobra.Command {
	initCmd := &cobra.Command{
		Use:   "init",
		Short: "初始化配置文件",
		Long: `在用户配置目录中创建新的配置文件。
配置文件将创建在以下位置：
	 - Windows: %USERPROFILE%\.config\ask\config.json
	 - macOS/Linux: ~/.config/ask/config.json

如果配置文件已存在，此命令将显示错误。`,
		Run: func(cmd *cobra.Command, args []string) {
			err := config.InitConfig()
			if err != nil {
				fmt.Printf("❌ 初始化配置失败: %s\n", err)
				os.Exit(1)
			}
			
			configPath := config.GetConfigPath()
			fmt.Printf("✅ 配置文件已成功创建: %s\n", configPath)
			fmt.Println("\n📝 请编辑配置文件，设置您的 API 密钥和其他设置。")
			fmt.Println("💡 您也可以通过环境变量设置配置:")
			fmt.Println("   ASK_API_URL - API 服务器地址")
			fmt.Println("   ASK_API_KEY - API 密钥")
		},
	}

	return initCmd
}