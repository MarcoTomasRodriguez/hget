package cmd

import (
	"context"
	"github.com/MarcoTomasRodriguez/hget/internal/download"
	"github.com/MarcoTomasRodriguez/hget/pkg/ctxutil"
	"github.com/MarcoTomasRodriguez/hget/pkg/logger"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"path/filepath"
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
		logger := logger.NewConsoleLogger()
		fs := afero.NewBasePathFs(afero.NewOsFs(), viper.GetString(ProgramFolderKey))
		ctx := ctxutil.NewCancelableContext(context.Background())

		// Initialize manager.
		manager := download.NewManager(fs)

		// Read download file.
		download, err := manager.GetDownloadById(args[0])

		// Start download.
		if err = manager.StartDownload(download, ctx); err != nil {
			logger.Error(err.Error())
			return
		}

		// Move download to output folder.
		if err := os.Rename(filepath.Join(viper.GetString("download_folder"), download.Id, download.Name), download.Name); err != nil {
			logger.Error(err.Error())
			return
		}

		// Delete program download folder.
		if err := manager.DeleteDownloadById(download.Id); err != nil {
			logger.Error(err.Error())
			return
		}
	},
}

// init registers the resume command.
func init() {
	rootCmd.AddCommand(resumeCmd)
}
