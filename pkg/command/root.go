package command

import (
	"context"
	"os"
	"strings"

	"github.com/owncloud/ocis/ocis-pkg/sync"

	"github.com/micro/cli/v2"
	"github.com/owncloud/ocis-wopiserver/pkg/config"
	"github.com/owncloud/ocis-wopiserver/pkg/flagset"
	"github.com/owncloud/ocis-wopiserver/pkg/version"
	"github.com/owncloud/ocis/ocis-pkg/log"
	"github.com/spf13/viper"
)

// Execute is the entry point for the ocis-wopiserver command.
func Execute(cfg *config.Config) error {
	app := &cli.App{
		Name:     "wopiserver",
		Version:  version.String,
		Usage:    "wopiserver, an example oCIS extension",
		Compiled: version.Compiled(),

		Authors: []*cli.Author{
			{
				Name:  "ownCloud GmbH",
				Email: "support@owncloud.com",
			},
		},

		Flags: flagset.RootWithConfig(cfg),

		Before: func(c *cli.Context) error {
			return ParseConfig(c, cfg)
		},

		Commands: []*cli.Command{
			Server(cfg),
			Health(cfg),
		},
	}

	cli.HelpFlag = &cli.BoolFlag{
		Name:  "help,h",
		Usage: "Show the help",
	}

	cli.VersionFlag = &cli.BoolFlag{
		Name:  "version,v",
		Usage: "Print the version",
	}

	return app.Run(os.Args)
}

// NewLogger initializes a service-specific logger instance.
func NewLogger(cfg *config.Config) log.Logger {
	return log.NewLogger(
		log.Name("wopiserver"),
		log.Level(cfg.Log.Level),
		log.Pretty(cfg.Log.Pretty),
		log.Color(cfg.Log.Color),
		log.File(cfg.Log.File),
	)
}

// ParseConfig loads wopiserver configuration from Viper known paths.
func ParseConfig(c *cli.Context, cfg *config.Config) error {
	sync.ParsingViperConfig.Lock()
	defer sync.ParsingViperConfig.Unlock()
	logger := NewLogger(cfg)

	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.SetEnvPrefix("WOPISERVER")
	viper.AutomaticEnv()

	if c.IsSet("config-file") {
		viper.SetConfigFile(c.String("config-file"))
	} else {
		viper.SetConfigName("wopiserver")

		viper.AddConfigPath("/etc/ocis")
		viper.AddConfigPath("$HOME/.ocis")
		viper.AddConfigPath("./config")
	}

	if err := viper.ReadInConfig(); err != nil {
		switch err.(type) {
		case viper.ConfigFileNotFoundError:
			logger.Info().
				Msg("Continue without config")
		case viper.UnsupportedConfigError:
			logger.Fatal().
				Err(err).
				Msg("Unsupported config type")
		default:
			logger.Fatal().
				Err(err).
				Msg("Failed to read config")
		}
	}

	if err := viper.Unmarshal(&cfg); err != nil {
		logger.Fatal().
			Err(err).
			Msg("Failed to parse config")
	}

	return nil
}

// SutureService allows for the wopiserver command to be embedded and supervised by a suture supervisor tree.
type SutureService struct {
	cfg *config.Config
}

// NewSutureService creates a new wopiserver.SutureService
// TODO: how to do this for non core extensions?
//func NewSutureService(cfg *ociscfg.Config) suture.Service {
//	if cfg.Mode == 0 {
//		cfg.WopiServer.Supervised = true
//	}
//	cfg.WopiServer.Log.File = cfg.Log.File
//	return SutureService{
//		cfg: cfg.WopiServer,
//	}
//}

func (s SutureService) Serve(ctx context.Context) error {
	s.cfg.Context = ctx
	if err := Execute(s.cfg); err != nil {
		return err
	}

	return nil
}
