package cmd

import (
	"github.com/MarcoTomasRodriguez/hget/download"
	"github.com/MarcoTomasRodriguez/hget/logger"
	"github.com/spf13/cobra"
)

// listCmd represents the list command.
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all the saved downloads.",
	Long: `List all the saved downloads.

For example:
$ hget list
INFO: Saved downloads:
 ⁕  9218d55b6ba5da11-go1.17.2.src.tar.gz  ⇒  URL: https://golang.org/dl/go1.17.2.src.tar.gz Size: 21.2 MB
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
			logger.LogInfo("There are no saved downloads.")
			return
		}

		// List the saved downloads.
		outputMessage := "Saved downloads:\n"
		for _, d := range downloads {
			outputMessage += d.String()
		}
		logger.LogInfo(outputMessage)
	},
}

// init registers the list command.
func init() {
	rootCmd.AddCommand(listCmd)
}
