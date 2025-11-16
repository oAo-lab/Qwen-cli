package commands

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"Qwen-cli/client"
	"Qwen-cli/config"
)

func TestCommand(cfg config.Config) *cobra.Command {
	return &cobra.Command{
		Use:   "test",
		Short: "测试模型或端点连接",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Testing connectivity to the model...")

			params := struct {
				Model    string `json:"model"`
				Messages []struct {
					Role    string `json:"role"`
					Content string `json:"content"`
				} `json:"messages"`
			}{
				Model: cfg.Models["default"].Name,
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

			err := client.Client(cfg.APIURL, cfg.APIKey, jsonParams, func(data []byte) {
				fmt.Println("连接测试成功！")
				var response struct {
					Choices []struct {
						Message struct {
							Content string `json:"content"`
						} `json:"message"`
					} `json:"choices"`
				}

				err := json.Unmarshal(data, &response)
				if err != nil {
					fmt.Printf("解析响应错误: %s\n", err)
					return
				}

				if len(response.Choices) > 0 {
					fmt.Println("模型响应:")
					fmt.Println(response.Choices[0].Message.Content)
				} else {
					fmt.Println("模型无响应。")
				}
			})

			if err != nil {
				fmt.Printf("错误: %s\n", err)
			}
		},
	}
}
