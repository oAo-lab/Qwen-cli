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
