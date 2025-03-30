package commands

import (
	"fmt"

	"github.com/spf13/cobra"

	"Qwen-cli/config"
)

var DEBUG = false

func DebugCommand(_ config.Config) *cobra.Command {
	return &cobra.Command{
		Use:   "debug",
		Short: "Set debug mode",
		Run: func(cmd *cobra.Command, args []string) {
			DEBUG = !DEBUG
			if DEBUG {
				fmt.Println("Debug mode enabled.")
			} else {
				fmt.Println("Debug mode disabled.")
			}
		},
	}
}
