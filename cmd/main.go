package main

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/spf13/cobra"

	"Qwen-cli/commands"
	"Qwen-cli/config"
	"Qwen-cli/utils"
)

func main() {
	exePath, err := os.Executable()
	if err != nil {
		fmt.Printf("Error getting executable path: %s\n", err)
		os.Exit(1)
	}

	exePath, err = filepath.Abs(exePath)
	if err != nil {
		fmt.Printf("Error resolving absolute path: %s\n", err)
		os.Exit(1)
	}

	exeDir := filepath.Dir(exePath)
	configPath := filepath.Join(exeDir, "config.json")

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		fmt.Printf("Error loading config: %s\n", err)
		os.Exit(1)
	}

	utils.DebugPrintln("Executable Path: " + exePath)
	utils.DebugPrintf("Config Path: %s\n", configPath)

	rootCmd := &cobra.Command{Use: "app"}
	rootCmd.AddCommand(commands.ChatCommand(cfg))
	rootCmd.AddCommand(commands.TestCommand(cfg))
	rootCmd.AddCommand(commands.DebugCommand(cfg))
	rootCmd.AddCommand(commands.CompletionCommand(rootCmd))

	// Handle SIGINT signal to pause the conversation
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT)

	go func() {
		for range signalChan {
			fmt.Println("\nConversation paused. Press Enter to continue...")
			fmt.Scanln()
		}
	}()

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
