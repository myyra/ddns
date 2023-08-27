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

const (
	ttlSeconds = 60
)

func main() {
	zoneFlag := flag.String("zoneName", "", "Zone name/domain")
	recordFlag := flag.String("recordName", "", "Record name/FQDN")
	tokenFlag := flag.String("token", "", "Cloudflare API token")
	intervalFlag := flag.Duration("interval", 1*time.Minute, "Update interval in minutes")
	debugFlag := flag.Bool("debug", false, "Use debug logging")
	flag.Parse()

	zoneName := *zoneFlag
	recordName := *recordFlag
	token := *tokenFlag
	if token == "" {
		token = os.Getenv("CLOUDFLARE_TOKEN")
	}
	interval := *intervalFlag
	debug := *debugFlag

	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	log := zerolog.New(os.Stdout).With().Logger()
	ctx := log.WithContext(context.Background())

	api, err := cloudflare.NewWithAPIToken(token)
	if err != nil {
		log.Fatal().Err(err).Msg("constructing Cloudflare API client")
	}

	updatesCtx, cancel := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer cancel()

	startUpdating(updatesCtx, api, zoneName, recordName, interval)
}

func startUpdating(
	ctx context.Context,
	api *cloudflare.API,
	zoneName, recordName string,
	interval time.Duration,
) {
	log := zerolog.Ctx(ctx)

	log.Info().Msg("running setup")
	var record *cloudflare.DNSRecord
	var zone *cloudflare.ResourceContainer
	for {
		log := log.With().Str("task", "setup").Logger()
		setupCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()
		setupCtx = log.WithContext(setupCtx)

		var err error
		record, zone, err = setup(setupCtx, api, zoneName, recordName)
		if err != nil {
			log.Error().Err(err).Msg("setting up initial record")
			log.Info().Msg("retrying in 10 minutes")

			select {
			case <-time.After(1 * time.Minute):
			case <-ctx.Done():
				log.Info().Msg("context done, exiting")
				return
			}

			continue
		}
		log.Info().Msg("setup complete")
		break
	}

	log.Info().Msg("starting update loop")
	ticker := time.NewTicker(interval)
	for {
		select {
		case <-ticker.C:
			log := log.With().Str("task", "update").Logger()
			updateCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
			updateCtx = log.WithContext(updateCtx)
			defer cancel()

			log.Info().Msg("checking IP")
			ip, err := getIP(updateCtx)
			if err != nil {
				log.Error().Err(err).Msg("getting IP")
				continue
			}

			if record.Content != ip {
				log.Info().Msg("IP changed, updating record")

				*record, err = update(updateCtx, api, zone, *record, ip)
				if err != nil {
					log.Error().Err(err).Msg("updating record")
					continue
				}
				log.Info().Msg("record updated")
			}
		case <-ctx.Done():
			log.Info().Msg("context done, exiting")
			return
		}
	}
}

func setup(ctx context.Context, api *cloudflare.API, zoneName, recordName string) (*cloudflare.DNSRecord, *cloudflare.ResourceContainer, error) {
	log := zerolog.Ctx(ctx).With().Logger()

	log.Debug().Msg("getting zone")
	zone, err := getZone(api, zoneName)
	if err != nil {
		return nil, nil, fmt.Errorf("getting zone container: %w", err)
	}

	log.Debug().Msg("getting record")
	record, err := firstRecordByName(ctx, api, zone, recordName)
	if err != nil {
		return nil, nil, fmt.Errorf("getting record: %w", err)
	}

	if record == nil {
		log.Info().Msg("record not found, creating")

		ip, err := getIP(ctx)
		if err != nil {
			return nil, nil, fmt.Errorf("getting IP: %w", err)
		}

		record, err = createRecord(ctx, api, zone, recordName, ip)
		if err != nil {
			return nil, nil, fmt.Errorf("create record: %w", err)
		}
	}

	return record, zone, nil
}
