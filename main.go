package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cloudflare/cloudflare-go"
	"github.com/rs/zerolog"
)

func main() {
	zoneFlag := flag.String("zoneName", "", "Zone name/domain")
	recordFlag := flag.String("recordName", "", "Record name/FQDN")
	tokenFlag := flag.String("token", "", "Cloudflare API token")
	flag.Parse()

	log := zerolog.New(os.Stdout).With().Str("zone_name", *zoneFlag).Str("record_name", *recordFlag).Logger()
	ctx := log.WithContext(context.Background())

	token := *tokenFlag
	if token == "" {
		token = os.Getenv("CLOUDFLARE_TOKEN")
	}

	api, err := cloudflare.NewWithAPIToken(token)
	if err != nil {
		log.Fatal().Err(err).Msg("constructing Cloudflare API client")
	}

	log.Info().Msg("doing initial update")
	err = update(ctx, api, *zoneFlag, *recordFlag)
	if err != nil {
		log.Error().Err(err).Msg("initial update")
	}

	log.Info().Msg("starting continuous updates")
	updatesCtx, _ := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	updatesDone := startUpdating(updatesCtx, api, *zoneFlag, *recordFlag)

	<-updatesDone
	os.Exit(0)
}

func startUpdating(
	ctx context.Context,
	api *cloudflare.API,
	zoneName, recordName string,
) chan struct{} {
	log := zerolog.Ctx(ctx)
	done := make(chan struct{})
	ticker := time.NewTicker(2 * time.Minute)
	go func() {
		defer ticker.Stop()
		defer close(done)
		for {
			select {
			case <-ctx.Done():
				log.Info().Msg("stopping updates")
				done <- struct{}{}
				return
			case <-ticker.C:
				ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
				defer cancel()
				err := update(ctx, api, zoneName, recordName)
				if err != nil {
					log.Error().Err(err).Msg("doing update")
				}
			}
		}
	}()
	return done
}

func update(ctx context.Context, api *cloudflare.API, zoneName, recordName string) error {
	log := zerolog.Ctx(ctx)
	ip, err := getIP(ctx)
	if err != nil {
		return fmt.Errorf("getting IP: %w", err)
	}

	if ctx.Err() != nil {
		return fmt.Errorf("context is done: %w", ctx.Err())
	}

	ctx = log.With().Str("current_ip", ip).Logger().WithContext(ctx)
	err = updateRecord(ctx, api, zoneName, recordName, ip)
	if err != nil {
		return fmt.Errorf("updating record: %w", err)
	}

	return nil
}
