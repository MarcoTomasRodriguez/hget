package cmd

import (
	"errors"

	"github.com/MarcoTomasRodriguez/hget/internal/download"
	"github.com/MarcoTomasRodriguez/hget/pkg/console"
	"github.com/MarcoTomasRodriguez/hget/pkg/logger"

	"github.com/spf13/cobra"
)

// resumeCmd represents the resume command.
var resumeCmd = &cobra.Command{
	Use:   "resume ID",
	Short: "Resumes a saved download.",
	Long: `Remove a saved download.

For example:
$ hget resume 01cc0f0a3d94af18-file1.txt`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Create application context.
		ctx := console.CancelableContext()

		// Read download file.
		d, err := download.GetDownload(args[0])

		if err == nil {
			if err := d.Execute(ctx); err != nil {
				logger.Error("An error occurred while downloading: %v", err)
			}
		} else if errors.Is(err, download.ErrDownloadNotExist) {
			logger.Error("Download does not exist.")
		} else if errors.Is(err, download.ErrDownloadBroken) {
			logger.Error("Download is broken, and thus will be removed.")

			if err := d.Delete(); err != nil {
				logger.Error("Could not delete saved download: %v", err)
			}
		} else {
			logger.Error("Unknown error: %v", err)
		}
	},
}

// init registers the resume command.
func init() {
	rootCmd.AddCommand(resumeCmd)
}
