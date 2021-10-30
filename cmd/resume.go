package cmd

import (
	"errors"

	"github.com/MarcoTomasRodriguez/hget/download"
	"github.com/MarcoTomasRodriguez/hget/logger"
	"github.com/MarcoTomasRodriguez/hget/utils"

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
		ctx := utils.ConsoleCancelableContext()

		// Read download file.
		d, err := download.GetDownload(args[0])

		if err == nil {
			if err := d.Execute(ctx); err != nil {
				logger.LogError("An error ocurred while downloading: %v", err)
			}
		} else if errors.Is(err, utils.ErrDownloadNotExist) {
			logger.LogError("Download does not exist.")
		} else if errors.Is(err, utils.ErrDownloadBroken) {
			logger.LogError("Download is broken, and thus will be removed.")

			if err := download.DeleteDownload(args[0]); err != nil {
				logger.LogError("Could not delete saved download: %v", err)
			}
		} else {
			logger.LogError("Unknown error: %v", err)
		}
	},
}

// init registers the resume command.
func init() {
	rootCmd.AddCommand(resumeCmd)
}
