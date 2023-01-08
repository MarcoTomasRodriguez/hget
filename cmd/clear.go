package cmd

import (
	"github.com/MarcoTomasRodriguez/hget/internal/download"
	"github.com/MarcoTomasRodriguez/hget/pkg/codec"
	"github.com/MarcoTomasRodriguez/hget/pkg/logger"
	"github.com/MarcoTomasRodriguez/hget/pkg/progressbar"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
		// Initialize downloader.
		logger := logger.NewConsoleLogger()
		fs := afero.NewBasePathFs(afero.NewOsFs(), viper.GetString("download_folder"))
		storage := download.NewStorage(fs, codec.NewYAMLCodec())
		downloader := download.NewDownloader(download.NewNetwork(), storage, progressbar.NewProgressBar(), logger)

		// List downloads.
		downloads, err := downloader.FindAllDownloads()
		if err != nil {
			logger.Error("Could not list downloads: %v", err)
			return
		}

		// Check if there are no saved downloads.
		if len(downloads) == 0 {
			logger.Info("There are no downloads to remove.")
			return
		}

		// List the removed downloads.
		outputMessage := "Removed downloads:\n"
		for _, d := range downloads {
			if err := downloader.DeleteDownloadById(d.Id); err != nil {
				logger.Error("Could not delete download: %v\n", err)
				continue
			}

			outputMessage += d.String()
		}
		logger.Info(outputMessage)
	},
}

// init registers the clear command.
func init() {
	rootCmd.AddCommand(clearCmd)
}
