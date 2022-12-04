package cmd

import (
	"github.com/MarcoTomasRodriguez/hget/internal/download"
	"github.com/MarcoTomasRodriguez/hget/pkg/logger"
	"github.com/samber/lo"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"strings"
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
		logger := logger.NewConsoleLogger()
		fs := afero.NewBasePathFs(afero.NewOsFs(), viper.GetString(ProgramFolderKey))

		// Initialize manager.
		manager := download.NewManager(fs)

		// List downloads.
		downloads, err := manager.ListDownloads()
		if err != nil {
			logger.Error("Could not list downloads: %v", err)
			return
		}

		// Check if there are no saved downloads.
		if len(downloads) == 0 {
			logger.Info("There are no saved downloads.")
			return
		}

		// List the saved downloads.
		downloadsString := lo.Map(downloads, func(d *download.Download, _ int) string {
			return d.String()
		})

		logger.Info("Saved downloads:\n" + strings.Join(downloadsString, ""))
	},
}

// init registers the list command.
func init() {
	rootCmd.AddCommand(listCmd)
}
