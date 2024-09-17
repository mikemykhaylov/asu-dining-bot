package cmd

import (
	"context"
	"errors"

	"github.com/mikemykhaylov/asu-dining-bot/internal"
	"github.com/mikemykhaylov/asu-dining-bot/internal/api"
	"github.com/mikemykhaylov/asu-dining-bot/internal/config"
	"github.com/mikemykhaylov/asu-dining-bot/internal/handler"
	"github.com/mikemykhaylov/asu-dining-bot/internal/logger"
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

			runHandler := handler.NewRunHandler()

			return runHandler.Run(ctx)
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
	if err := viper.BindEnv(config.PersonalIDKey, "PERSONAL_ID"); err != nil {
		panic(err)
	}
	viper.SetDefault(config.PersonalIDKey, api.TelegramPublicPersonalID)

	// bool flag for running in docker
	serveCmd.Flags().BoolP("docker", "", false, "Set to true when running in Docker. Sets chromium to /usr/bin/chromium-browser")
	if err := viper.BindPFlag(config.DockerKey, serveCmd.Flags().Lookup("docker")); err != nil {
		panic(err)
	}
	viper.SetDefault(config.DockerKey, false)

	rootCmd.AddCommand(serveCmd)
}

func Execute() error {
	return rootCmd.Execute()
}
