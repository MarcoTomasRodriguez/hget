package cmd

import (
	"github.com/MarcoTomasRodriguez/hget/internal/download"
	"github.com/MarcoTomasRodriguez/hget/pkg/logger"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
		logger := logger.NewConsoleLogger()
		fs := afero.NewBasePathFs(afero.NewOsFs(), viper.GetString(DownloadFolderKey))

		// Initialize manager.
		manager := download.NewManager(fs)

		// Delete download using first command line argument as id.
		if err := manager.DeleteDownloadById(args[0]); err != nil {
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
