package cmd

import (
	"errors"
	"github.com/MarcoTomasRodriguez/hget/internal/download"
	"github.com/MarcoTomasRodriguez/hget/pkg/logger"
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
		d, err := download.GetDownload(args[0])

		// Check if download does not exist.
		if errors.Is(err, download.ErrDownloadNotExist) {
			logger.Error("Download does not exist.")
			return
		}

		// Remove download.
		if err := d.Delete(); err != nil {
			logger.Error("Could not remove download: %v", err)
			return
		}

		logger.Info("Download removed successfully.")
	},
}

// init adds removeCmd to rootCmd.
func init() {
	rootCmd.AddCommand(removeCmd)
}
