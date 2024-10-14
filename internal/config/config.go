package config

type BrowserMode string

var (
	Host   BrowserMode = "host"
	Docker BrowserMode = "docker"
	Remote BrowserMode = "remote"

	TelegramBotTokenKey string = "token"
	PersonalIDKey       string = "personalID"
	BrowserModeKey      string = "browserMode"
	ServerKey           string = "server"
)

type ServerConfig struct {
	Port uint16
}
