package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sf-news-comments/internal/comments"
	"sf-news-comments/internal/config"
	"sf-news-comments/internal/http_server"
	"sf-news-comments/internal/http_server/handlers"
	"sf-news-comments/pkg/logger"
	"syscall"
	"time"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := config.NewConfig()

	lgr, err := logger.NewLogger(os.Stdout, cfg.LogLevel)
	if err != nil {
		log.Fatalln(err)
	}

	lgr = lgr.With().
		CallerWithSkipFrameCount(2).
		Str("app", "sf-news-comments").
		Logger()

	shutdownCh := make(chan os.Signal, 1)
	signal.Notify(shutdownCh, syscall.SIGKILL, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	comments := comments.NewComments(cfg, lgr)

	handler := handlers.NewHandler(cfg, lgr, comments)
	httpServer, listenHTTPErr := http_server.NewServer(cfg, lgr, handler)

mainLoop:
	for {
		select {
		case <-ctx.Done():
			break mainLoop

		case err = <-listenHTTPErr:
			if err != nil {
				lgr.Error().Err(err).Msg("http server error")
				shutdownCh <- syscall.SIGTERM
			}

		case sig := <-shutdownCh:
			lgr.Info().Msgf("shutdown signal received: %s", sig.String())

			if err = httpServer.Shutdown(); err != nil {
				lgr.Error().Err(err).Msg("shutdown http server error")
			}

			if err = comments.Shutdown(); err != nil {
				lgr.Error().Err(err).Msg("shutdown comments enrichment error")
			}

			lgr.Info().Msg("server loop stopped")
			cancel()
			time.Sleep(1 * time.Second)
		}
	}
}
