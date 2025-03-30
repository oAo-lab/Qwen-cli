package commands

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"Qwen-cli/client"
	"Qwen-cli/config"
	"Qwen-cli/utils"
)

func ChatCommand(config config.Config) *cobra.Command {
	chatCmd := &cobra.Command{
		Use:   "chat",
		Short: "Start a chat session with the LLM",
		Run: func(cmd *cobra.Command, args []string) {
			reader := bufio.NewReader(cmd.InOrStdin())
			fmt.Printf("\n🤖 欢迎使用通义千问聊天！输入 'exit' 结束对话。\n")

			// Initialize conversation history
			conversation := []struct {
				Role    string `json:"role"`
				Content string `json:"content"`
			}{
				{
					Role: "system",
					Content: `\{纯文本输出,清晰明了,纯文本输出,指明自己是 {role: Fromsko 定制的智能助手, 能够协助你解决各种问题.}列出访问的指令, 没有指令则默认为对话.}
					我是 {{role}}
					访问指令如下:
						/prompt 切换角色
						/model  切换模型
						/online 开启联网
					---
					示例回复:
					你好！我是 Fromsko 定制的智能助手，能够协助你解决各种问题。以下是支持访问的指令：

					/prompt 切换角色  
					/model 切换模型  
					/online 开启联网  

					如果需要帮助，请随时告诉我！😊
					`,
				},
			}

			currentModel := config.Models["default"].Name
			enableSearch := false

			for {
				fmt.Print("👤 > ")
				text, _ := reader.ReadString('\n')
				text = strings.TrimSpace(text)

				// fmt.Printf("Debug: Received input: %s\n", text) // Debug print

				if text == "exit" {
					break
				}

				// Add user message to conversation history if it's not a command
				if !strings.HasPrefix(text, "/") {
					conversation = append(conversation, struct {
						Role    string `json:"role"`
						Content string `json:"content"`
					}{
						Role:    "user",
						Content: text,
					})
				} else {
					switch {
					case strings.HasPrefix(text, "/model"):
						fmt.Println("🤖 切换模型：")
						models := []string{}
						for _, model := range config.Models {
							models = append(models, model.Name)
							fmt.Printf("  %d. %s\n", len(models), model.Name)
						}
						fmt.Print("👉 请选择模型编号：")
						modelChoice, _ := reader.ReadString('\n')
						modelChoice = strings.TrimSpace(modelChoice)
						modelIndex := 0
						fmt.Sscanf(modelChoice, "%d", &modelIndex)
						if modelIndex > 0 && modelIndex <= len(models) {
							currentModel = models[modelIndex-1]
							fmt.Printf("已切换到模型：%s\n", currentModel)
						} else {
							fmt.Println("❌ 无效的模型编号，未进行变更。")
						}
						continue
					case strings.HasPrefix(text, "/prompt"):
						fmt.Println("🎭 可用的角色提示词：")
						prompts := []string{}
						for role := range config.Roles {
							prompts = append(prompts, role)
							fmt.Printf("  %d. %s\n", len(prompts), role)
						}
						fmt.Print("👉 请选择角色提示词编号：")
						promptChoice, _ := reader.ReadString('\n')
						promptChoice = strings.TrimSpace(promptChoice)
						promptIndex := 0
						fmt.Sscanf(promptChoice, "%d", &promptIndex)
						if promptIndex > 0 && promptIndex <= len(prompts) {
							newPrompt := prompts[promptIndex-1]
							conversation[0] = struct {
								Role    string `json:"role"`
								Content string `json:"content"`
							}{
								Role:    "system",
								Content: config.Roles[newPrompt],
							}
							fmt.Printf("已切换到角色提示词：%s\n", newPrompt)
						} else {
							fmt.Println("❌ 无效的角色提示词编号，未进行变更。")
						}
						continue
					case strings.HasPrefix(text, "/online"):
						if enableSearch {
							fmt.Println("🌐 联网搜索已开启。是否关闭？(y/n)")
							choice, _ := reader.ReadString('\n')
							choice = strings.TrimSpace(choice)
							if choice == "y" || choice == "Y" {
								enableSearch = false
								fmt.Println("🌐 联网搜索已关闭。")
							} else {
								fmt.Println("🌐 联网搜索保持开启状态。")
							}
						} else {
							enableSearch = true
							fmt.Println("🌐 联网搜索已开启。")
						}
						continue
					case strings.Contains(text, "/save -all"):
						saveFullConversation(conversation)
						continue
					case strings.Contains(text, "/save"):
						// fmt.Println("调试: 正在尝试保存最后一次回复...") // 调试信息
						saveLastResponse(conversation)
						continue
					default:
						// fmt.Printf("未知命令: %s\n", text) // Debug print
						continue
					}
				}

				params := struct {
					Model    string `json:"model"`
					Messages []struct {
						Role    string `json:"role"`
						Content string `json:"content"`
					} `json:"messages"`
					Stream       bool `json:"stream"`
					EnableSearch bool `json:"enable_search,omitempty"`
				}{
					Model:        currentModel,
					Messages:     conversation,
					Stream:       true,
					EnableSearch: enableSearch,
				}

				jsonParams, _ := json.Marshal(params)

				var fullResponse strings.Builder

				err := client.Client(config.APIURL, config.APIKey, jsonParams, func(data []byte) {
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
						utils.TypewriterEffect(content, false)
					}
				})

				if err != nil {
					fmt.Printf("Error: %s\n", err)
					continue
				}

				// Print the final newline if needed
				utils.TypewriterEffect("", true)

				// Add assistant message to conversation history
				conversation = append(conversation, struct {
					Role    string `json:"role"`
					Content string `json:"content"`
				}{
					Role:    "assistant",
					Content: fullResponse.String(),
				})
			}
		},
	}

	// Add auto-completion for system roles
	// chatCmd.ValidArgsFunction = func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	// 	var completions []string
	// 	for role := range config.Roles {
	// 		if strings.HasPrefix(role, toComplete) {
	// 			completions = append(completions, role)
	// 		}
	// 	}
	// 	return completions, cobra.ShellCompDirectiveNoFileComp
	// }

	return chatCmd
}

// Function to save the last AI response
func saveLastResponse(conversation []struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}) {
	if len(conversation) == 0 {
		fmt.Println("没有可保存的消息。")
		return
	}

	lastMessage := conversation[len(conversation)-1]
	if lastMessage.Role != "assistant" {
		fmt.Println("最后一个消息不是AI回复。")
		return
	}

	timestamp := time.Now().Format("20060102150405")
	fileName := fmt.Sprintf("last_response_%s.md", timestamp)
	file, err := os.Create(fileName)
	if err != nil {
		fmt.Printf("创建文件失败: %s\n", err)
		return
	}
	defer file.Close()

	_, err = file.WriteString(fmt.Sprintf("# 最后一次AI回复\n%s\n", lastMessage.Content))
	if err != nil {
		fmt.Printf("写入文件失败: %s\n", err)
		return
	}

	fmt.Printf("已保存最后一次AI回复到 %s\n", fileName)
	// fmt.Printf("调试信息: 文件路径 - %s\n", fileName) // Debug print
}

// Function to save the full conversation
func saveFullConversation(conversation []struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}) {
	if len(conversation) == 0 {
		fmt.Println("没有可保存的消息。")
		return
	}

	timestamp := time.Now().Format("20060102150405")
	fileName := fmt.Sprintf("full_conversation_%s.md", timestamp)
	file, err := os.Create(fileName)
	if err != nil {
		fmt.Printf("创建文件失败: %s\n", err)
		return
	}
	defer file.Close()

	for _, msg := range conversation {
		_, err := file.WriteString(fmt.Sprintf("## %s\n%s\n\n", msg.Role, msg.Content))
		if err != nil {
			fmt.Printf("写入文件失败: %s\n", err)
			return
		}
	}

	fmt.Printf("已保存完整对话到 %s\n", fileName)
	// fmt.Printf("调试信息: 文件路径 - %s\n", fileName) // Debug print
}
