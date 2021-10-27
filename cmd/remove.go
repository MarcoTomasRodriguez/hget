package cmd

import (
	"github.com/MarcoTomasRodriguez/hget/download"
	"github.com/MarcoTomasRodriguez/hget/logger"
	"github.com/spf13/cobra"
)

// removeCmd represents the remove command.
var removeCmd = &cobra.Command{
	Use:   "remove ID",
	Short: "Remove a saved download.",
	Long: `Remove a saved download.

For example:
$ hget remove 01cc0f0a3d94af18-file1.txt
INFO: download removed successfully.
`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Remove download.
		if err := download.DeleteDownload(args[0]); err != nil {
			logger.LogError("Could not remove download: %v", err)
			return
		}

		logger.LogInfo("Download removed successfully.")
	},
}

// init adds removeCmd to rootCmd.
func init() {
	rootCmd.AddCommand(removeCmd)
}
