package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/alecthomas/kong"
	"github.com/goschtalt/casemapper"
	"github.com/goschtalt/goschtalt"
	_ "github.com/goschtalt/yaml-decoder"
	_ "github.com/goschtalt/yaml-encoder"
	"github.com/mitchellh/mapstructure"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/schmidtw/heaticus-maximus/httpserver"
	"github.com/schmidtw/heaticus-maximus/views"
	"github.com/xmidt-org/sallust"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	applicationName = "heaticus-maximus"
)

var (
	commit  = "undefined"
	version = "undefined"
	date    = "undefined"
	builtBy = "undefined"
)

type CLI struct {
	Debug bool     `optional:"" help:"Run in debug mode."`
	Show  bool     `optional:"" short:"s" help:"Show the configuration and exit."`
	Files []string `optional:"" short:"f" name:"file" help:"Specific configuration files."`
	Dirs  []string `optional:"" short:"d" name:"dir" help:"Specific configuration directories."`
}

func heaticus(args []string) (exitCode int) {
	app := fx.New(
		fx.WithLogger(func(log *zap.Logger) fxevent.Logger {
			return &fxevent.ZapLogger{Logger: log}
		}),
		fx.Provide(
			// Handle the CLI processing and return the processed input.
			func() *CLI {
				var cli CLI
				_ = kong.Parse(&cli,
					kong.Name(applicationName),
					kong.Description("A house heater controller.\n"+
						fmt.Sprintf("\tVersion:  %s\n", version)+
						fmt.Sprintf("\tDate:     %s\n", date)+
						fmt.Sprintf("\tCommit:   %s\n", commit)+
						fmt.Sprintf("\tBuilt By: %s\n", builtBy),
					),
					kong.UsageOnError(),
				)
				return &cli
			},

			// Collect and process the configuration files and env vars and
			// produce a configuration object.
			func(cli *CLI) (*goschtalt.Config, error) {
				return goschtalt.New(
					goschtalt.AutoCompile(),
					goschtalt.ExpandEnv(),
					goschtalt.DefaultMarshalOptions(
						goschtalt.IncludeOrigins(),
						goschtalt.FormatAs("yml"),
					),
					goschtalt.DefaultUnmarshalOptions(
						casemapper.ConfigStoredAs("two_words"),
						goschtalt.DecodeHook(
							mapstructure.StringToTimeDurationHookFunc(),
						),
					),
					goschtalt.AddJumbled(os.DirFS("/"), os.DirFS("."),
						append(cli.Files, cli.Dirs...)...),
				)
			},

			// Create the logger and configure it based on if the program is in
			// debug mode or normal mode.
			goschtalt.UnmarshalFn[sallust.Config]("logger", goschtalt.Optional()),
			func(cli *CLI, cfg sallust.Config) (*zap.Logger, error) {
				if cli.Debug {
					cfg.Level = "DEBUG"
					cfg.Development = true
					cfg.Encoding = "console"
					cfg.EncoderConfig = sallust.EncoderConfig{
						TimeKey:        "T",
						LevelKey:       "L",
						NameKey:        "N",
						CallerKey:      "C",
						FunctionKey:    zapcore.OmitKey,
						MessageKey:     "M",
						StacktraceKey:  "S",
						LineEnding:     zapcore.DefaultLineEnding,
						EncodeLevel:    "capitalColor",
						EncodeTime:     "RFC3339",
						EncodeDuration: "string",
						EncodeCaller:   "short",
					}
					cfg.OutputPaths = []string{"stderr"}
					cfg.ErrorOutputPaths = []string{"stderr"}
				}
				return cfg.Build()
			},

			// Define the metrics endpoint based on it's configuration.
			fx.Annotate(
				goschtalt.UnmarshalFn[httpserver.Config]("servers.metrics"),
				fx.ResultTags(`name:"metrics.config"`),
			),
			fx.Annotate(
				promhttp.Handler,
				fx.ResultTags(`name:"metrics.handler"`),
			),
			fx.Annotate(
				httpserver.New,
				fx.ParamTags(`name:""`, `name:"metrics.handler"`, `name:"metrics.config"`),
				fx.ResultTags(`name:"metrics.server"`),
			),

			// Define the ui endpoint based on it's configuration.
			fx.Annotate(
				goschtalt.UnmarshalFn[httpserver.Config]("servers.ui"),
				fx.ResultTags(`name:"ui.config"`),
			),
			fx.Annotate(
				views.Handler,
				fx.ResultTags(`name:"ui.handler"`),
			),
			fx.Annotate(
				httpserver.New,
				fx.ParamTags(`name:""`, `name:"ui.handler"`, `name:"ui.config"`),
				fx.ResultTags(`name:"ui.server"`),
			),
		),
		fx.Invoke(
			// Require the metrics server to start.
			fx.Annotate(
				func(*http.Server) {},
				fx.ParamTags(`name:"metrics.server"`),
			),

			// Require the ui server to start.
			fx.Annotate(
				func(*http.Server) {},
				fx.ParamTags(`name:"ui.server"`),
			),

			// Handle the -s/--show option where the configuration is shown,
			// then the program is exited.
			func(cli *CLI, cfg *goschtalt.Config) {
				if cli.Show {
					fmt.Fprintln(os.Stdout, cfg.Explain())

					out, err := cfg.Marshal()
					if err != nil {
						fmt.Fprintln(os.Stderr, err)
					} else {
						fmt.Fprintln(os.Stdout, "---\n"+string(out))
					}
					// Exit so extra errors aren't output.
					os.Exit(0)
				}
			},
		),
	)

	switch err := app.Err(); {
	case err == nil:
		app.Run()
	default:
		fmt.Fprintln(os.Stderr, err)
		return 2
	}

	return 0
}

func main() {
	os.Exit(heaticus(os.Args))
}
