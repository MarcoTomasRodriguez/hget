package cmd

import (
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

		d, err := download.GetDownload(args[0])

		switch err {
		case utils.ErrDownloadNotExist:
			logger.LogError("download does not exist.\n")
		case utils.ErrDownloadBroken:
			logger.LogError("download is broken, and thus will be removed.\n")
			_ = download.DeleteDownload(args[0])
		default:
			err := d.Execute(ctx)
			logger.LogError(err.Error())
		}
	},
}

// init registers the resume command.
func init() {
	rootCmd.AddCommand(resumeCmd)
}
