package cmd

import (
	"fmt"
	"github.com/bilalcaliskan/syn-flood/internal/logging"
	"github.com/bilalcaliskan/syn-flood/internal/options"
	"github.com/bilalcaliskan/syn-flood/internal/raw"
	"github.com/bilalcaliskan/syn-flood/internal/version"
	"github.com/dimiro1/banner"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"os"
	"strings"
	"time"
)

var (
	opts   *options.SynFloodOptions
	ver    = version.Get()
	logger = logging.GetLogger()
)

// init function initializes the cmd module
func init() {
	opts = options.GetSynFloodOptions()
	// 213.238.175.187
	rootCmd.PersistentFlags().StringVarP(&opts.Host, "host", "",
		"213.238.175.187", "Provide public ip or DNS of the target")
	rootCmd.PersistentFlags().IntVarP(&opts.Port, "port", "", 443,
		"Provide reachable port of the target")
	rootCmd.PersistentFlags().IntVarP(&opts.PayloadLength, "payloadLength", "",
		1400, "Provide payload length in bytes for each SYN packet")
	rootCmd.PersistentFlags().StringVarP(&opts.FloodType, "floodType", "", raw.TypeSyn,
		"Provide the attack type. Proper values are: syn, ack, synack")
	rootCmd.PersistentFlags().Int64VarP(&opts.FloodDurationSeconds, "floodDurationSeconds",
		"", -1, "Provide the duration of the attack in seconds, -1 for no limit, defaults to -1")
	rootCmd.Flags().StringVarP(&opts.BannerFilePath, "bannerFilePath", "", "build/ci/banner.txt",
		"relative path of the banner file")
	rootCmd.Flags().BoolVarP(&opts.VerboseLog, "verbose", "v", false, "verbose output of the logging library (default false)")

	if err := rootCmd.Flags().MarkHidden("bannerFilePath"); err != nil {
		logger.Fatal("fatal error occured while hiding flag", zap.Error(err))
	}
}

var (
	// rootCmd represents the base command when called without any subcommands
	rootCmd = &cobra.Command{
		Use:     "syn-flood",
		Short:   "A simple flooding tool written with Golang",
		Version: ver.GitVersion,
		Long: `This project is developed with the objective of learning low level network
operations with Golang. It starts a syn flood attack with raw sockets.
Please do not use that tool with devil needs.
`,
		Run: func(cmd *cobra.Command, args []string) {
			if opts.VerboseLog {
				logging.Atomic.SetLevel(zap.DebugLevel)
			}

			shouldStop := make(chan bool)
			go func() {
				if err = raw.StartFlooding(shouldStop, opts.Host, opts.Port, opts.PayloadLength, opts.FloodType); err != nil {
					fmt.Println()
					logger.Error("an error occurred on flooding process", zap.String("error", err.Error()))
					os.Exit(1)
				}
			}()

			go func() {
				if opts.FloodDurationSeconds != -1 {
					<-time.After(time.Duration(opts.FloodDurationSeconds) * time.Second)
					shouldStop <- true
					close(shouldStop)
				}
			}()

			for {
				select {
				case <-shouldStop:
					logger.Info("shouldStop channel received a signal, stopping")
					return
				default:
					continue
				}
			}
		},
	}
	err error
)

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	bannerBytes, _ := os.ReadFile("banner.txt")
	banner.Init(os.Stdout, true, false, strings.NewReader(string(bannerBytes)))

	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
