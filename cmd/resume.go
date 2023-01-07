package cmd

import (
	"context"
	"github.com/MarcoTomasRodriguez/hget/internal/download"
	"github.com/MarcoTomasRodriguez/hget/pkg/codec"
	"github.com/MarcoTomasRodriguez/hget/pkg/ctxutil"
	"github.com/MarcoTomasRodriguez/hget/pkg/logger"
	"github.com/MarcoTomasRodriguez/hget/pkg/progressbar"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"path/filepath"
)

// resumeCmd represents the resume command.
var resumeCmd = &cobra.Command{
	Use:   "resume ID",
	Short: "Resumes a saved _download.",
	Long: `Remove a saved _download.

For example:
$ hget resume 01cc0f0a3d94af18-file1.txt`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Initialize downloader.
		logger := logger.NewConsoleLogger()
		fs := afero.NewBasePathFs(afero.NewOsFs(), viper.GetString("download_folder"))
		storage := download.NewStorage(fs, codec.NewYAMLCodec())
		downloader := download.NewDownloader(download.NewNetwork(), storage, progressbar.NewProgressBar(), logger)

		// Read download specification.
		download, err := downloader.FindDownloadById(args[0])

		// Start download.
		ctx := ctxutil.NewCancelableContext(context.Background())
		if err = downloader.Download(download, ctx); err != nil {
			logger.Error(err.Error())
			return
		}

		// Move download to output folder.
		if err := os.Rename(filepath.Join(viper.GetString("download_folder"), download.Id, "output"), download.Name); err != nil {
			logger.Error(err.Error())
			return
		}

		// Delete program _download folder.
		if err := downloader.DeleteDownloadById(download.Id); err != nil {
			logger.Error(err.Error())
			return
		}
	},
}

// init registers the resume command.
func init() {
	rootCmd.AddCommand(resumeCmd)
}
