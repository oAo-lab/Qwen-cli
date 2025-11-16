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
		Short: "Initialize configuration file",
		Long: `Initialize a new configuration file in the user's config directory.
The configuration file will be created at:
  - Windows: %USERPROFILE%\.config\ask\config.json
  - macOS/Linux: ~/.config/ask/config.json

If the configuration file already exists, this command will show an error.`,
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