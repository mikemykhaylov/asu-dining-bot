package cmd

import (
	"context"
	"errors"

	"github.com/mikemykhaylov/asu-dining-bot/internal"
	"github.com/mikemykhaylov/asu-dining-bot/internal/api"
	"github.com/mikemykhaylov/asu-dining-bot/internal/config"
	"github.com/mikemykhaylov/asu-dining-bot/internal/handler"
	"github.com/mikemykhaylov/asu-dining-bot/internal/logger"
	"github.com/mikemykhaylov/asu-dining-bot/internal/server"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	rootCmd = &cobra.Command{
		Use:     "asu-dining-bot",
		Short:   "ASU Dining Bot is a bot that gets today's Barrett dinner menu from the ASU dining website and sends them to you via Telegram",
		Version: internal.Version(),
	}
	serveCmd = &cobra.Command{
		Use:   "run",
		Short: "Run the bot",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			log := logger.FromContext(ctx)
			log.Info("Running bot")

			token := viper.GetString("token")
			if token == "" {
				log.Error("Telegram bot token is required")
				return errors.New("telegram bot token is required")
			}

			personalID := viper.GetInt64("personalID")
			if personalID == api.TelegramPublicPersonalID {
				log.Error("Telegram personal ID is required")
				return errors.New("telegram personal ID is required")
			}

			mode := viper.GetString(config.BrowserModeKey)
			if mode != string(config.Host) && mode != string(config.Docker) && mode != string(config.Remote) {
				log.Error("Invalid browser mode", "mode", mode)
				return errors.New("invalid browser mode")
			}

			port := viper.GetUint16("port")
			runAsServer := viper.GetBool(config.ServerKey)
			if runAsServer {
				log.Info("Running in server mode", "port", port)
				config := &config.ServerConfig{
					Port: port,
				}

				err := server.NewServer(config)
				if err != nil {
					log.Error("Failed to run server", "cause", err)
				}
				return nil
			}

			runHandler := handler.NewRunHandler()

			err := runHandler.Run(ctx)
			if err != nil {
				log.Error("Failed to run handler", "cause", err)
			}
			return nil
		},
	}
)

func init() {
	serveCmd.Flags().StringP("token", "t", "", "Telegram bot token")
	if err := viper.BindPFlag(config.TelegramBotTokenKey, serveCmd.Flags().Lookup("token")); err != nil {
		panic(err)
	}
	if err := viper.BindEnv(config.TelegramBotTokenKey, "TELEGRAM_BOT_TOKEN"); err != nil {
		panic(err)
	}

	serveCmd.Flags().StringP("personalID", "", "", "Telegram personal ID")
	if err := viper.BindPFlag(config.PersonalIDKey, serveCmd.Flags().Lookup("personalID")); err != nil {
		panic(err)
	}
	if err := viper.BindEnv(config.PersonalIDKey, "TELEGRAM_PERSONAL_ID"); err != nil {
		panic(err)
	}
	viper.SetDefault(config.PersonalIDKey, api.TelegramPublicPersonalID)

	serveCmd.Flags().StringP("browserMode", "", "host", "Browser mode: host, docker, or remote")
	if err := viper.BindPFlag(config.BrowserModeKey, serveCmd.Flags().Lookup("browserMode")); err != nil {
		panic(err)
	}
	if err := viper.BindEnv(config.BrowserModeKey, "BROWSER_MODE"); err != nil {
		panic(err)
	}
	viper.SetDefault(config.BrowserModeKey, string(config.Host))

	serveCmd.Flags().BoolP("server", "", false, "Run as a server")
	if err := viper.BindPFlag(config.ServerKey, serveCmd.Flags().Lookup("server")); err != nil {
		panic(err)
	}
	if err := viper.BindEnv(config.ServerKey, "AS_SERVER"); err != nil {
		panic(err)
	}
	viper.SetDefault(config.ServerKey, false)

	serveCmd.Flags().Uint16P("port", "p", 8080, "Port to listen on")
	if err := viper.BindPFlag("port", serveCmd.Flags().Lookup("port")); err != nil {
		panic(err)
	}
	if err := viper.BindEnv("port", "PORT"); err != nil {
		panic(err)
	}
	viper.SetDefault("port", 8080)

	rootCmd.AddCommand(serveCmd)
}

func Execute() error {
	return rootCmd.Execute()
}
