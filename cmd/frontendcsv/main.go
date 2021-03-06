package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/atb-as/kindly/cmd/frontendcsv/http"
	"github.com/atb-as/kindly/statistics"
	"github.com/atb-as/kindly/statistics/auth"
	"github.com/go-kit/kit/log"
	"golang.org/x/oauth2"
)

type config struct {
	listenPort string
	botID      string
	apiKey     string
}

func main() {
	listenPortFlag := flag.String("port", "8080", "HTTP listen port")
	botIDFlag := flag.String("botid", "", "kindly bot ID")
	apiKeyFlag := flag.String("apikey", "", "kindly API key")
	flag.Parse()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	if err := run(ctx, &config{
		listenPort: *listenPortFlag,
		botID:      *botIDFlag,
		apiKey:     *apiKeyFlag,
	}); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		os.Exit(1)
	}
}

func run(ctx context.Context, config *config) error {
	client := statistics.NewClient(
		statistics.WithDoer(oauth2.NewClient(context.Background(), oauth2.ReuseTokenSource(nil, &auth.TokenSource{
			APIKey: config.apiKey,
			BotID:  config.botID,
		}))),
		statistics.WithLogger(log.NewLogfmtLogger(os.Stdout)))
	client.BotID = config.botID

	srv := http.NewServer(client, config.listenPort)

	go func() {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			fmt.Fprintf(os.Stderr, "srv.ListenAndServe: err=%v\n", err)
		}
	}()

	<-ctx.Done()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return srv.Shutdown(ctx)

}
