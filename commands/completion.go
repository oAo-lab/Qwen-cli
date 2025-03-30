package commands

import (
	"os"

	"github.com/spf13/cobra"
)

func CompletionCommand(rootCmd *cobra.Command) *cobra.Command {
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
