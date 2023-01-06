package cmd

import (
	"context"
	"github.com/MarcoTomasRodriguez/hget/internal/download"
	"github.com/MarcoTomasRodriguez/hget/pkg/ctxutil"
	"github.com/MarcoTomasRodriguez/hget/pkg/progressbar"
	"github.com/spf13/afero"
	"math/rand"
	"time"

	"github.com/MarcoTomasRodriguez/hget/pkg/logger"
	"os"
	"path/filepath"
	"runtime"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	ProgramFolderKey  = "program_folder"
	DownloadFolderKey = "download_folder"
)

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:   "hget URL",
	Short: "Interruptible and resumable _download accelerator",
	Long: `Interruptible and resumable _download accelerator.

hget allows you to _download at the maximum speed possible using
_download threads and to stop and resume tasks.
`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		logger := logger.NewConsoleLogger()
		fs := afero.NewBasePathFs(afero.NewOsFs(), viper.GetString("download_folder"))
		downloader := download.NewDownloader(download.NewNetwork(), download.NewStorage(fs), progressbar.NewProgressBar(), logger)

		// Get number of workers from flags.
		workers, _ := cmd.Flags().GetUint8("workers")

		// Load download from url.
		download, err := downloader.InitDownload(args[0], workers)
		if err != nil {
			logger.Error(err.Error())
			return
		}

		// Start download.
		ctx := ctxutil.NewCancelableContext(context.Background())
		if err := downloader.Download(download, ctx); err != nil {
			logger.Error(err.Error())
			return
		}

		// Move download to output folder.
		if err := os.Rename(filepath.Join(viper.GetString("download_folder"), download.Id, "output"), download.Name); err != nil {
			logger.Error(err.Error())
			return
		}

		// Delete internal download folder.
		if err := downloader.DeleteDownloadById(download.Id); err != nil {
			logger.Error(err.Error())
			return
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	// Seed math/rand.
	rand.Seed(time.Now().UnixNano())

	// Define program folder global flag.
	homeDir, _ := os.UserHomeDir()
	defaultProgramFolder := filepath.Join(homeDir, ".hget")
	rootCmd.PersistentFlags().String(ProgramFolderKey, defaultProgramFolder, "Configures the program folder.")
	_ = viper.BindPFlag(ProgramFolderKey, rootCmd.PersistentFlags().Lookup(ProgramFolderKey))

	// Define _download folder global flag.
	defaultDownloadFolder := filepath.Join(homeDir, ".hget/downloads")
	rootCmd.PersistentFlags().String(DownloadFolderKey, defaultDownloadFolder, "Configures the _download folder.")
	_ = viper.BindPFlag(DownloadFolderKey, rootCmd.PersistentFlags().Lookup(DownloadFolderKey))

	// Define log level global flag.
	rootCmd.PersistentFlags().Int("log", 2, "Set log level: 0 means no logs, 1 only important logs and 2 all logs.")
	_ = viper.BindPFlag("log_level", rootCmd.PersistentFlags().Lookup("log"))

	// Define worker numbers flag.
	rootCmd.Flags().Uint8P("workers", "n", uint8(runtime.NumCPU()), "Set number of _download workers.")

	// Create internal download folder.
	_ = afero.NewOsFs().MkdirAll(viper.GetString("download_folder"), 0755)
}
