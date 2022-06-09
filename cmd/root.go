package cmd

import (
	"context"
	"github.com/MarcoTomasRodriguez/hget/internal/config"
	"github.com/samber/do"
	"github.com/spf13/afero"
	"log"
	"math/rand"
	"time"

	"github.com/MarcoTomasRodriguez/hget/internal/download"
	"github.com/MarcoTomasRodriguez/hget/pkg/console"
	"github.com/MarcoTomasRodriguez/hget/pkg/logger"

	"os"
	"path/filepath"
	"runtime"

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
		ctx := console.CancelableContext(context.Background())

		// Get number of workers from flags.
		// TODO: Add worker limit.
		workers, err := cmd.Flags().GetInt("workers")
		if err != nil {
			logger.Error("Could not get number of workers from flags.")
			return
		}

		// Start download.
		d, err := download.NewDownload(args[0], workers)
		if err != nil {
			logger.Error("Could not start download: %v", err)
			return
		}

		if err := d.Execute(ctx); err != nil {
			logger.Error("An error occurred while downloading: %v", err)
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func initializeDependencies() {
	configPath, _ := rootCmd.PersistentFlags().GetString("config_path")
	do.ProvideValue[*config.Config](nil, config.NewConfig(configPath))
	do.ProvideValue[*log.Logger](nil, log.New(os.Stdout, "", 0))
	do.ProvideValue[*afero.Afero](nil, &afero.Afero{Fs: afero.NewOsFs()})
}

func init() {
	// Seed math/rand.
	rand.Seed(time.Now().UnixNano())

	// Define config global flag.
	homeDir, _ := os.UserHomeDir()
	configPath := filepath.Join(homeDir, ".hget/config.toml")
	rootCmd.PersistentFlags().String("config_path", configPath, "Set the configuration path.")

	// Define log level global flag.
	rootCmd.PersistentFlags().Int("log", 2, "Set log level: 0 means no logs, 1 only important logs and 2 all logs.")
	_ = viper.BindPFlag("log_level", rootCmd.PersistentFlags().Lookup("log"))

	// Define worker numbers flag.
	rootCmd.Flags().IntP("workers", "n", runtime.NumCPU(), "Set number of download workers.")

	// Initialize dependencies.
	cobra.OnInitialize(initializeDependencies)
}
