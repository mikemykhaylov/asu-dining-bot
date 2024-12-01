package server

import (
	"fmt"
	"net/http"

	"github.com/mikemykhaylov/asu-dining-bot/internal/config"
	"github.com/mikemykhaylov/asu-dining-bot/internal/handler"
	"github.com/mikemykhaylov/asu-dining-bot/internal/logger"
)

func requestHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)
	runHandler := handler.NewRunHandler()

	var err error

	// try up to 5 times to get the menu
	for i := 0; i < 5; i++ {
		err = runHandler.Run(ctx)
		if err != nil {
			log.Error(fmt.Sprintf("failed to get the menu, attempt %d", i+1), "cause", err)
		} else {
			break
		}
	}

	if err != nil {
		log.Error("Gave up trying to get the menu after 5 attempts")
	}

	fmt.Fprint(w, "OK")
}

func NewServer(serverConfig *config.ServerConfig) error {
	http.HandleFunc("/", requestHandler)
	address := fmt.Sprintf("0.0.0.0:%d", serverConfig.Port)

	return http.ListenAndServe(address, nil)
}
