package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var (
	DEBUG   = false
	VERSION = "0.1.3"
)

type ModelConfig struct {
	Name string `json:"name"`
}

type Config struct {
	APIURL string                 `json:"api_url"`
	APIKey string                 `json:"api_key"`
	Models map[string]ModelConfig `json:"models"`
	Roles  map[string]string      `json:"roles"`
}

func loadConfig(filename string) (Config, error) {
	var config Config
	file, err := os.Open(filename)
	if err != nil {
		return config, err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		return config, err
	}
	return config, nil
}

func client(apiURL, apiKey string, params []byte, callBack func(data []byte)) error {
	reader := bytes.NewReader(params)

	req, err := http.NewRequest("POST", apiURL, reader)
	if err != nil {
		return fmt.Errorf("error creating request: %s", err.Error())
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("error sending request: %s", err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("error reading error response: %s, status code: %d", err.Error(), resp.StatusCode)
		}
		return fmt.Errorf("API error: %s, status code: %d", string(bodyBytes), resp.StatusCode)
	}

	// Handle streaming response
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "data:") {
			data := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
			if data == "[DONE]" {
				break
			}
			callBack([]byte(data))
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading stream: %s", err.Error())
	}

	return nil
}

func typewriterEffect(text string, done bool) {
	fmt.Print(text)
	if done {
		fmt.Println()
	}
}

func chatCommand(config Config) *cobra.Command {
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

				err := client(config.APIURL, config.APIKey, jsonParams, func(data []byte) {
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
						typewriterEffect(content, false)
					}
				})

				if err != nil {
					fmt.Printf("Error: %s\n", err)
					continue
				}

				// Print the final newline if needed
				typewriterEffect("", true)

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
	chatCmd.ValidArgsFunction = func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		var completions []string
		for role := range config.Roles {
			if strings.HasPrefix(role, toComplete) {
				completions = append(completions, role)
			}
		}
		return completions, cobra.ShellCompDirectiveNoFileComp
	}

	return chatCmd
}

func testCommand(config Config) *cobra.Command {
	return &cobra.Command{
		Use:   "test",
		Short: "Test the model or endpoint connectivity",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Testing connectivity to the model...")

			params := struct {
				Model    string `json:"model"`
				Messages []struct {
					Role    string `json:"role"`
					Content string `json:"content"`
				} `json:"messages"`
			}{
				Model: config.Models["default"].Name,
				Messages: []struct {
					Role    string `json:"role"`
					Content string `json:"content"`
				}{
					{
						Role:    "system",
						Content: "You are a helpful assistant.",
					},
					{
						Role:    "user",
						Content: "Hello",
					},
				},
			}

			jsonParams, _ := json.Marshal(params)

			err := client(config.APIURL, config.APIKey, jsonParams, func(data []byte) {
				fmt.Println("Connectivity test successful!")
				var response struct {
					Choices []struct {
						Message struct {
							Content string `json:"content"`
						} `json:"message"`
					} `json:"choices"`
				}

				err := json.Unmarshal(data, &response)
				if err != nil {
					fmt.Printf("Error parsing response: %s\n", err)
					return
				}

				if len(response.Choices) > 0 {
					fmt.Println("Response from model:")
					fmt.Println(response.Choices[0].Message.Content)
				} else {
					fmt.Println("No response from model.")
				}
			})

			if err != nil {
				fmt.Printf("Error: %s\n", err)
			}
		},
	}
}

func debugCommand(_ Config) *cobra.Command {
	return &cobra.Command{
		Use:   "debug",
		Short: "set debug mode",
		Run: func(cmd *cobra.Command, args []string) {
			DEBUG = !DEBUG
		},
	}
}

func completionCommand(rootCmd *cobra.Command) *cobra.Command {
	return &cobra.Command{
		Use:   "completion [bash|zsh]",
		Short: "Generate completion script",
		Long: `To load completions:

Bash:

$ source <(your-program completion bash)

# To load completions for each session, execute once:
Linux:
  $ your-program completion bash > /etc/bash_completion.d/your-program
MacOS:
  $ your-program completion bash > /usr/local/etc/bash_completion.d/your-program

Zsh:

# If shell completion is not already enabled in your environment you will need
# to enable it. You can execute the following once:

$ echo "autoload -U compinit; compinit" >> ~/.zshrc

# To load completions for each session, execute once:
$ your-program completion zsh > "${fpath[1]}/_your-program"

# You will need to start a new shell for this setup to take effect.
`,
		DisableFlagsInUseLine: true,
		ValidArgs:             []string{"bash", "zsh"},
		Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		Run: func(cmd *cobra.Command, args []string) {
			switch args[0] {
			case "bash":
				rootCmd.GenBashCompletion(os.Stdout)
			case "zsh":
				rootCmd.GenZshCompletion(os.Stdout)
			}
		},
	}
}

func debugPrintln(e ...any) {
	if DEBUG {
		fmt.Println(e...)
	}
}

func debugPrintf(f string, e ...any) {
	if DEBUG {
		fmt.Printf(f, e...)
	}
}

func main() {
	exePath, err := os.Executable()
	if err != nil {
		fmt.Printf("Error getting executable path: %s\n", err)
		os.Exit(1)
	}

	// Resolve absolute path
	exePath, err = filepath.Abs(exePath)
	if err != nil {
		fmt.Printf("Error resolving absolute path: %s\n", err)
		os.Exit(1)
	}

	exeDir := filepath.Dir(exePath)
	configPath := filepath.Join(exeDir, "config.json")

	config, err := loadConfig(configPath)
	if err != nil {
		fmt.Printf("Error loading config: %s\n", err)
		os.Exit(1)
	}

	debugPrintln("Executable Path: " + exePath)
	debugPrintf("Config Path: %s\n", configPath)

	rootCmd := &cobra.Command{Use: "app"}
	rootCmd.AddCommand(chatCommand(config))
	rootCmd.AddCommand(testCommand(config))
	rootCmd.AddCommand(debugCommand(config))
	rootCmd.AddCommand(completionCommand(rootCmd))

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
