package commands

import (
	"fmt"
	"runtime"

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
	var force bool

	updateCmd := &cobra.Command{
		Use:   "update",
		Short: "检查并更新到最新版本",
		Long:  `检查GitHub上是否有新版本，如果有则提示用户是否更新。使用 --force 参数可以强制更新到最新版本。`,
		Run: func(cmd *cobra.Command, args []string) {
			if force {
				fmt.Println("🔥 强制更新模式：将更新到最新版本")
				performForceUpdate()
				return
			}

			fmt.Println("� 正在检查更新...")

			hasUpdate, release, err := version.CheckUpdate()
			if err != nil {
				fmt.Printf("❌ 检查更新失败: %s\n", err)
				return
			}

			if !hasUpdate {
				fmt.Printf("✅ 您使用的是最新版本: %s\n", version.GetVersion())
				fmt.Println("💡 提示: 使用 'ask update --force' 可以强制重新安装最新版本")
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

	// 添加强制更新标志
	updateCmd.Flags().BoolVarP(&force, "force", "f", false, "强制更新到最新版本，即使当前已是最新版本")

	return updateCmd
}

// performForceUpdate 执行强制更新
func performForceUpdate() {
	fmt.Println("🔍 正在获取最新版本信息...")

	release, err := version.GetLatestRelease()
	if err != nil {
		fmt.Printf("❌ 获取最新版本失败: %s\n", err)
		return
	}

	if release == nil {
		fmt.Println("❌ 获取最新版本失败: 返回数据为空")
		return
	}

	currentVersion := version.GetVersion()
	fmt.Printf("📋 当前版本: %s\n", currentVersion)
	fmt.Printf("🎯 目标版本: %s\n", release.TagName)

	if currentVersion == release.TagName {
		fmt.Println("ℹ️  当前版本已是最新，但将强制重新安装...")
	}

	// 获取下载URL
	downloadURL := version.GetDownloadURL(release)
	if downloadURL == "" {
		fmt.Println("❌ 无法找到适合您系统的下载文件")
		fmt.Printf("🔍 调试信息: 系统=%s, 架构=%s\n", runtime.GOOS, runtime.GOARCH)
		fmt.Printf("🔍 可用资源文件:\n")
		for i, asset := range release.Assets {
			fmt.Printf("  %d. %s\n", i+1, asset.Name)
		}
		return
	}

	fmt.Printf("🔗 下载地址: %s\n", downloadURL)

	// 询问用户是否确认强制更新
	fmt.Print("⚠️  强制更新将重新安装当前版本，是否继续? (y/N): ")
	var confirm string
	fmt.Scanln(&confirm)

	if confirm != "y" && confirm != "Y" && confirm != "yes" && confirm != "YES" {
		fmt.Println("❌ 已取消强制更新")
		return
	}

	fmt.Println("🚀 正在下载并强制安装更新...")

	err = version.DownloadAndInstall(downloadURL)
	if err != nil {
		fmt.Printf("❌ 强制更新失败: %s\n", err)
		return
	}

	fmt.Println("✅ 强制更新完成! 请重新启动程序以使用新版本")
}
