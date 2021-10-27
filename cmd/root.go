package cmd

import (
	"errors"

	"github.com/MarcoTomasRodriguez/hget/config"
	"github.com/MarcoTomasRodriguez/hget/download"
	"github.com/MarcoTomasRodriguez/hget/logger"

	// "github.com/MarcoTomasRodriguez/hget/logger"
	"os"
	"path/filepath"
	"runtime"

	"github.com/MarcoTomasRodriguez/hget/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:   "hget URL",
	Short: "Interruptible and resumable download accelerator",
	Long: `Interruptible and resumable download accelerator.

hget allows you to download at the maximum speed possible using
download threads and to stop and resume tasks.
`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Create application context.
		ctx := utils.ConsoleCancelableContext()

		workers, _ := cmd.Flags().GetUint16("workers")
		downloadID, _ := download.ParseDownloadID(args[0])
		d, err := download.GetDownload(downloadID)

		if errors.Is(err, utils.ErrDownloadBroken) {
			logger.LogError("Download is broken, and thus will be removed.")
			_ = download.DeleteDownload(downloadID)
		}

		if d == nil {
			d, _ = download.NewDownload(args[0], workers)
		}

		d.Execute(ctx)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	// Initialize config.
	cobra.OnInitialize(config.LoadConfig)

	// Define config global flag.
	homeDir, _ := os.UserHomeDir()
	// logger.CheckErr(err)
	rootCmd.PersistentFlags().StringVar(&config.Filepath, "config", filepath.Join(homeDir, ".hget/config.toml"), "Set config file.")

	// Define log level global flag.
	rootCmd.PersistentFlags().Uint8("log", uint8(2), "Set log level: 0 means no logs, 1 only important logs and 2 all logs.")
	_ = viper.BindPFlag("log_level", rootCmd.PersistentFlags().Lookup("log"))
	// logger.CheckErr(err)

	// Define worker numbers flag.
	rootCmd.Flags().Uint16P("workers", "n", uint16(runtime.NumCPU()), "Set number of download workers.")
}
