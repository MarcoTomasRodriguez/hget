package cmd

import (
	"github.com/MarcoTomasRodriguez/hget/download"
	"github.com/MarcoTomasRodriguez/hget/logger"
	"github.com/spf13/cobra"
)

// clearCmd represents the clear command
var clearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Removes all the saved downloads.",
	Long: `Removes all the saved downloads.

For example:
$ hget clear
INFO: Removed downloads:
 ⁕  9218d55b6ba5da11-go1.17.2.src.tar.gz  ⇒  URL: https://golang.org/dl/go1.17.2.src.tar.gz size: 21.2 MB
`,
	Run: func(cmd *cobra.Command, args []string) {
		// List downloads.
		downloads, err := download.ListDownloads()
		if err != nil {
			logger.LogError("Could not list downloads: %v", err)
			return
		}

		// Check if there are no saved downloads.
		if len(downloads) == 0 {
			logger.LogInfo("There are no downloads to remove.\n")
			return
		}

		// List the removed downloads.
		outputMessage := "Removed downloads:\n"
		for _, d := range downloads {
			if err := download.DeleteDownload(d.ID); err != nil {
				logger.LogError("Could not delete download: %v\n", err)
				continue
			}

			outputMessage += d.String()
		}
		logger.LogInfo(outputMessage)
	},
}

// init registers the clear command.
func init() {
	rootCmd.AddCommand(clearCmd)
}
