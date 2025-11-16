package commands

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"

	"Qwen-cli/client"
	"Qwen-cli/config"
	"Qwen-cli/utils"
)

func CmdCommand(cfg config.Config) *cobra.Command {
	cmdCmd := &cobra.Command{
		Use:   "cmd",
		Short: "AI助手 - 支持普通聊天和命令执行",
		Long: `AI助手支持普通聊天和系统命令生成执行两种模式。
两种模式共享对话上下文，可以无缝切换。

使用方法：
	 ask cmd                    # 启动AI助手
	 ask cmd "描述您的需求"       # 直接描述需求，AI会生成命令

功能特性：
	 - 普通聊天：直接输入文本进行对话
	 - 命令模式：使用 /cmd 前缀生成并执行系统命令
	 - 上下文共享：两种模式共享对话历史`,
		Run: func(cmd *cobra.Command, args []string) {
			reader := bufio.NewReader(cmd.InOrStdin())
			
			// 初始化对话历史
			conversation := []struct {
				Role    string `json:"role"`
				Content string `json:"content"`
			}{
				{
					Role: "system",
					Content: `你是一个专业的系统命令助手。你的唯一任务是根据用户的需求生成合适的系统命令。

重要规则：
1. 只输出可执行的命令，不要任何解释、描述或回答
2. 用户已经通过 /cmd 前缀明确表示需要命令，所以所有输入都是命令请求
3. 确保命令安全，避免破坏性操作
4. 如果需要多个步骤，请使用 && 或 ; 连接
5. 根据操作系统选择合适的命令语法（Windows使用cmd或PowerShell，macOS/Linux使用bash）
6. 优先使用跨平台的命令
7. 如果用户需求不明确，请询问具体细节

示例：
用户：查看当前目录的文件
输出：ls -la

用户：创建一个名为test的目录
输出：mkdir test

用户：查看系统信息
输出：uname -a

用户：查看docker容器
输出：docker ps -a

用户：查看端口8080是否被占用
输出：lsof -i :8080

项目信息：
如果用户询问项目相关信息，请提供以下信息：
- 项目地址：https://github.com/oAo-lab/Qwen-cli
- 项目名称：Qwen-cli
- 项目描述：通义千问命令行客户端，支持多模型对话和角色切换`,
				},
			}

			currentModel := cfg.Models["default"].Name
			
			// 获取环境信息
			osInfo := utils.GetEnvironmentInfo()

			// 如果有参数，直接使用作为用户请求（非交互模式）
			if len(args) > 0 {
				userRequest := strings.Join(args, " ")
				
				// 添加用户请求到对话历史
				conversation = append(conversation, struct {
					Role    string `json:"role"`
					Content string `json:"content"`
				}{
					Role:    "user",
					Content: userRequest,
				})

				// 准备API请求
				params := struct {
					Model    string `json:"model"`
					Messages []struct {
						Role    string `json:"role"`
						Content string `json:"content"`
					} `json:"messages"`
					Stream bool `json:"stream"`
				}{
					Model:    currentModel,
					Messages: conversation,
					Stream:   true,
				}

				jsonParams, _ := json.Marshal(params)

				fmt.Printf("\n🤔 AI正在思考...\n")

				var fullResponse strings.Builder

				// 调用AI生成命令
				err := client.Client(cfg.APIURL, cfg.APIKey, jsonParams, func(data []byte) {
					var response struct {
						Choices []struct {
							Delta struct {
								Content string `json:"content"`
							} `json:"delta"`
						} `json:"choices"`
					}

					err := json.Unmarshal(data, &response)
					if err != nil {
						fmt.Printf("Error parsing response: %s\n", err)
						return
					}

					if len(response.Choices) > 0 {
						content := response.Choices[0].Delta.Content
						fullResponse.WriteString(content)
						// 流式显示AI生成的命令
						fmt.Print(content)
					}
				})

				if err != nil {
					fmt.Printf("❌ 错误: %s\n", err)
					return
				}

				// 显示生成的命令
				generatedCmd := strings.TrimSpace(fullResponse.String())
				fmt.Printf("\n\n💡 AI生成的命令：\n\n")
				fmt.Printf("```bash\n%s\n```\n\n", generatedCmd)

				// 检查是否是有效的命令
				if strings.Contains(generatedCmd, "请描述您想要执行的操作") ||
				   strings.HasPrefix(generatedCmd, "请") ||
				   len(generatedCmd) == 0 {
					fmt.Println("💡 这不是一个有效的命令，请重新描述您的需求。")
					return
				}

				// 确认执行
				fmt.Printf("⚠️  请确认是否执行此命令？(y/N): ")
				confirm, _ := reader.ReadString('\n')
				confirm = strings.TrimSpace(strings.ToLower(confirm))

				if confirm != "y" && confirm != "yes" {
					fmt.Println("❌ 已取消执行")
					return
				}

				// 执行命令并捕获输出
				fmt.Printf("\n🚀 正在执行命令...\n\n")
				
				// 使用shell执行命令，支持管道和重定向，并捕获输出
				execCmd := exec.Command("sh", "-c", generatedCmd)
				var out bytes.Buffer
				var stderr bytes.Buffer
				execCmd.Stdout = &out
				execCmd.Stderr = &stderr
				
				err = execCmd.Run()
				
				// 获取命令输出
				commandOutput := out.String()
				commandError := stderr.String()
				
				// 显示命令输出（流式显示）
				if commandOutput != "" {
					fmt.Print(commandOutput)
				}
				if commandError != "" {
					fmt.Print(commandError)
				}
				
				if err != nil {
					fmt.Printf("\n❌ 命令执行失败: %s\n", err)
				} else {
					fmt.Printf("\n✅ 命令执行完成\n")
				}
				return
			}

			// 交互模式
			fmt.Printf("\n🤖 欢迎使用AI助手！\n")
			fmt.Printf("💡 提示：输入 'exit' 退出，输入 'help' 查看示例\n")
			fmt.Printf("💡 支持两种模式：\n")
			fmt.Printf("   - 普通聊天：直接输入文本进行对话\n")
			fmt.Printf("   - 命令模式：使用 '/cmd 命令描述' 生成并执行系统命令\n")
			fmt.Printf("💡 两种模式共享对话上下文，可以无缝切换\n\n")
			
			// 交互模式循环
			for {
				fmt.Print("👤 > ")
				text, _ := reader.ReadString('\n')
				text = strings.TrimSpace(text)

				if text == "exit" {
					fmt.Println("👋 再见！")
					return
				}
				
				if text == "help" {
					fmt.Println("\n📚 使用方法：")
					fmt.Println("  普通聊天：直接输入文本，AI会回答您的问题")
					fmt.Println("  命令模式：/cmd 命令描述，AI会生成并执行系统命令")
					fmt.Println()
					fmt.Println("💡 特性：")
					fmt.Println("  - 两种模式共享对话上下文")
					fmt.Println("  - 可以在聊天和命令模式之间无缝切换")
					fmt.Println("  - AI会记住之前的对话内容")
					fmt.Println()
					fmt.Println("📚 命令示例：")
					fmt.Println("  /cmd 查看当前目录的文件")
					fmt.Println("  /cmd 创建一个名为test的目录")
					fmt.Println("  /cmd 查看系统信息")
					fmt.Println("  /cmd 查看端口8080是否被占用")
					fmt.Println("  /cmd 查看磁盘使用情况")
					fmt.Println("  /cmd 安装npm包")
					fmt.Println()
					fmt.Println("📚 聊天示例：")
					fmt.Println("  你好")
					fmt.Println("  解释一下什么是Docker")
					fmt.Println("  如何学习Go语言")
					fmt.Println()
					continue
				}
				
				if text == "" {
					fmt.Println("❌ 请输入内容")
					continue
				}
				
				// 检查是否是命令请求
				isCommandRequest := strings.HasPrefix(text, "/cmd ")
				var userRequest string
				
				if isCommandRequest {
					userRequest = strings.TrimSpace(strings.TrimPrefix(text, "/cmd "))
					if userRequest == "" {
						fmt.Println("❌ 请在 /cmd 后描述您想要执行的命令")
						continue
					}
				} else {
					// 普通聊天，使用用户输入作为请求
					userRequest = text
				}
				
				// 添加用户请求到对话历史
				conversation = append(conversation, struct {
					Role    string `json:"role"`
					Content string `json:"content"`
				}{
					Role:    "user",
					Content: userRequest,
				})

				// 更新系统提示词，包含环境信息
				if isCommandRequest {
					// 命令模式下的系统提示词
					conversation[0].Content = fmt.Sprintf(`你是一个专业的系统命令助手。你的任务是根据用户的需求生成合适的系统命令。

环境信息：
%s

当前模型：%s

重要规则：
1. 只输出可执行的命令，不要任何解释、描述或回答
2. 用户已经通过 /cmd 前缀明确表示需要命令，所以这个请求是命令请求
3. 确保命令安全，避免破坏性操作
4. 如果需要多个步骤，请使用 && 或 ; 连接
5. 根据操作系统选择合适的命令语法（Windows使用cmd或PowerShell，macOS/Linux使用bash）
6. 优先使用跨平台的命令
7. 如果用户需求不明确，请询问具体细节

示例：
用户：查看当前目录的文件
输出：ls -la

用户：创建一个名为test的目录
输出：mkdir test

用户：查看系统信息
输出：uname -a

用户：查看docker容器
输出：docker ps -a

用户：查看端口8080是否被占用
输出：lsof -i :8080

项目信息：
如果用户询问项目相关信息，请提供以下信息：
- 项目地址：https://github.com/oAo-lab/Qwen-cli
- 项目名称：Qwen-cli
- 项目描述：通义千问命令行客户端，支持多模型对话和角色切换`, osInfo, currentModel)
				} else {
					// 普通聊天模式下的系统提示词
					conversation[0].Content = fmt.Sprintf(`你是一个智能助手，可以帮助用户解答问题和执行系统命令。

环境信息：
%s

当前模型：%s

你的能力：
1. 回答用户的各种问题和咨询
2. 提供技术支持和建议
3. 如果用户需要执行系统命令，可以提供命令建议
4. 保持对话的上下文连贯性

项目信息：
如果用户询问项目相关信息，请提供以下信息：
- 项目地址：https://github.com/oAo-lab/Qwen-cli
- 项目名称：Qwen-cli
- 项目描述：通义千问命令行客户端，支持多模型对话和角色切换

请以友好、专业的方式与用户交流。`, osInfo, currentModel)
				}

				// 准备API请求
				params := struct {
					Model    string `json:"model"`
					Messages []struct {
						Role    string `json:"role"`
						Content string `json:"content"`
					} `json:"messages"`
					Stream bool `json:"stream"`
				}{
					Model:    currentModel,
					Messages: conversation,
					Stream:   true,
				}

				jsonParams, _ := json.Marshal(params)

				fmt.Printf("\n🤔 AI正在思考...\n")

				var fullResponse strings.Builder

				// 调用AI生成命令
				err := client.Client(cfg.APIURL, cfg.APIKey, jsonParams, func(data []byte) {
					var response struct {
						Choices []struct {
							Delta struct {
								Content string `json:"content"`
							} `json:"delta"`
						} `json:"choices"`
					}

					err := json.Unmarshal(data, &response)
					if err != nil {
						fmt.Printf("Error parsing response: %s\n", err)
						return
					}

					if len(response.Choices) > 0 {
						content := response.Choices[0].Delta.Content
						fullResponse.WriteString(content)
						// 流式显示AI响应
						fmt.Print(content)
					}
				})

				if err != nil {
					fmt.Printf("❌ 错误: %s\n", err)
					continue
				}

				// 获取AI响应
				aiResponse := strings.TrimSpace(fullResponse.String())
				
				if isCommandRequest {
					// 命令模式处理
					fmt.Printf("\n\n💡 AI生成的命令：\n\n")
					fmt.Printf("```bash\n%s\n```\n\n", aiResponse)

					// 检查是否是有效的命令
					if strings.Contains(aiResponse, "请描述您想要执行的操作") ||
					   strings.HasPrefix(aiResponse, "请") ||
					   len(aiResponse) == 0 {
						fmt.Println("💡 这不是一个有效的命令，请重新描述您的需求。")
						// 添加AI响应到对话历史
						conversation = append(conversation, struct {
							Role    string `json:"role"`
							Content string `json:"content"`
						}{
							Role:    "assistant",
							Content: aiResponse,
						})
						continue
					}

					// 确认执行
					fmt.Printf("⚠️  请确认是否执行此命令？(y/N): ")
					confirm, _ := reader.ReadString('\n')
					confirm = strings.TrimSpace(strings.ToLower(confirm))

					if confirm != "y" && confirm != "yes" {
						fmt.Println("❌ 已取消执行")
						// 添加AI响应到对话历史，即使没有执行
						conversation = append(conversation, struct {
							Role    string `json:"role"`
							Content string `json:"content"`
						}{
							Role:    "assistant",
							Content: aiResponse,
						})
						continue
					}

					// 执行命令并捕获输出
					fmt.Printf("\n🚀 正在执行命令...\n\n")
					
					// 使用shell执行命令，支持管道和重定向，并捕获输出
					execCmd := exec.Command("sh", "-c", aiResponse)
					var out bytes.Buffer
					var stderr bytes.Buffer
					execCmd.Stdout = &out
					execCmd.Stderr = &stderr
					
					err = execCmd.Run()
					
					// 获取命令输出
					commandOutput := out.String()
					commandError := stderr.String()
					
					// 显示命令输出（流式显示）
					if commandOutput != "" {
						fmt.Print(commandOutput)
					}
					if commandError != "" {
						fmt.Print(commandError)
					}
					
					if err != nil {
						fmt.Printf("\n❌ 命令执行失败: %s\n", err)
					} else {
						fmt.Printf("\n✅ 命令执行完成\n")
					}

					// 将命令和结果添加到对话历史中
					conversation = append(conversation, struct {
						Role    string `json:"role"`
						Content string `json:"content"`
					}{
						Role:    "assistant",
						Content: aiResponse,
					})
					
					// 添加命令执行结果到对话历史
					resultText := commandOutput
					if commandError != "" {
						if resultText != "" {
							resultText += "\n"
						}
						resultText += "错误输出: " + commandError
					}
					if err != nil {
						if resultText != "" {
							resultText += "\n"
						}
						resultText += "执行错误: " + err.Error()
					}
					
					conversation = append(conversation, struct {
						Role    string `json:"role"`
						Content string `json:"content"`
					}{
						Role:    "user",
						Content: "命令执行结果:\n" + resultText,
					})

					fmt.Printf("\n🔄 是否继续使用命令助手？(y/N): ")
					continueConfirm, _ := reader.ReadString('\n')
					continueConfirm = strings.TrimSpace(strings.ToLower(continueConfirm))
					
					if continueConfirm != "y" && continueConfirm != "yes" {
						return
					}
				} else {
					// 普通聊天模式处理
					fmt.Printf("\n") // 只添加换行，因为内容已经在流式显示中输出过了
					
					// 添加AI响应到对话历史
					conversation = append(conversation, struct {
						Role    string `json:"role"`
						Content string `json:"content"`
					}{
						Role:    "assistant",
						Content: aiResponse,
					})
				}
			}
		},
	}

	return cmdCmd
}
