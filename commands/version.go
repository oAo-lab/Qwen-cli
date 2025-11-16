package commands

import (
	"fmt"
	"github.com/spf13/cobra"

	"Qwen-cli/version"
)

// VersionCommand 创建版本命令
func VersionCommand() *cobra.Command {
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "显示版本信息",
		Long:  `显示Qwen-cli的当前版本信息，包括构建时间和Git提交信息。`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(version.GetVersionInfo())
		},
	}

	return versionCmd
}

// UpdateCommand 创建更新命令
func UpdateCommand() *cobra.Command {
	updateCmd := &cobra.Command{
		Use:   "update",
		Short: "检查并更新到最新版本",
		Long:  `检查GitHub上是否有新版本，如果有则提示用户是否更新。`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("🔍 正在检查更新...")
			
			hasUpdate, release, err := version.CheckUpdate()
			if err != nil {
				fmt.Printf("❌ 检查更新失败: %s\n", err)
				return
			}

			if !hasUpdate {
				fmt.Printf("✅ 您使用的是最新版本: %s\n", version.GetVersion())
				return
			}

			fmt.Printf("🎉 发现新版本: %s\n", release.TagName)
			fmt.Printf("📅 发布时间: %s\n", release.PublishedAt.Format("2006-01-02 15:04:05"))
			fmt.Printf("📝 更新说明:\n%s\n", release.Body)

			// 获取下载URL
			downloadURL := version.GetDownloadURL(release)
			if downloadURL == "" {
				fmt.Println("❌ 无法找到适合您系统的下载文件")
				return
			}

			fmt.Printf("🔗 下载地址: %s\n", downloadURL)

			// 询问用户是否更新
			fmt.Print("⚠️  是否立即更新? (y/N): ")
			var confirm string
			fmt.Scanln(&confirm)

			if confirm != "y" && confirm != "Y" && confirm != "yes" && confirm != "YES" {
				fmt.Println("❌ 已取消更新")
				return
			}

			fmt.Println("🚀 正在下载并安装更新...")
			
			err = version.DownloadAndInstall(downloadURL)
			if err != nil {
				fmt.Printf("❌ 更新失败: %s\n", err)
				return
			}

			fmt.Println("✅ 更新完成! 请重新启动程序以使用新版本")
		},
	}

	return updateCmd
}