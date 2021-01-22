package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cloudflare/cloudflare-go"
	"github.com/rs/zerolog/log"
)

func main() {
	zoneFlag := flag.String("zoneName", "", "Zone name/domain")
	recordFlag := flag.String("recordName", "", "Record name/FQDN")
	tokenFlag := flag.String("token", "", "Cloudflare API token")

	flag.Parse()

	api, err := cloudflare.NewWithAPIToken(*tokenFlag)
	if err != nil {
		log.Fatal().Err(err).Msg("constructing Cloudflare API client")
	}

	log.Info().Msg("starting DNS updates")

	go func() {
		logger := log.With().Str("zone_name", *zoneFlag).Str("record_name", *recordFlag).Logger()

		ip, err := getIP()
		if err != nil {
			logger.Error().Err(err).Str("func", "get_ip").Msg("error getting IP")
		}

		logger = logger.With().Str("current_ip", ip).Str("func", "update").Logger()

		err = updateRecord(logger, api, *zoneFlag, *recordFlag, ip)
		if err != nil {
			logger.Error().Err(err).Msg("error updating record")
		}

		time.Sleep(2 * time.Minute)
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	log.Info().Msg("stopping")
	os.Exit(0)
}
