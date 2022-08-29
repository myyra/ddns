package main

import (
	"context"
	"fmt"
	"time"

	"github.com/cloudflare/cloudflare-go"
	"github.com/rs/zerolog"
)

func updateRecord(
	ctx context.Context,
	api *cloudflare.API,
	zoneName string,
	recordName, ip string,
) error {
	log := zerolog.Ctx(ctx)
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	zoneID, err := api.ZoneIDByName(zoneName)
	if err != nil {
		return fmt.Errorf("finding zone by name %s: %w", zoneName, err)
	}
	newRecord := cloudflare.DNSRecord{
		Type:    "A",
		Name:    recordName,
		TTL:     120,
		Content: ip,
	}

	currentRecord, err := getCurrentRecord(ctx, api, zoneID, recordName)
	if err != nil {
		return fmt.Errorf("getting current record for zone %s: %w", zoneName, err)
	}

	if currentRecord == nil {
		log.Info().Msg("record not found, creating")

		resp, err := api.CreateDNSRecord(ctx, zoneID, newRecord)
		if err != nil {
			return fmt.Errorf("create record: %w", err)
		}
		if len(resp.Errors) != 0 {
			return fmt.Errorf("create record returned errors: %v", resp.Errors)
		}

		return nil
	}

	if currentRecord.Content != ip {
		log.Info().Str("old_ip", currentRecord.Content).Msg("IP has changed, updating")

		err := api.UpdateDNSRecord(ctx, zoneID, currentRecord.ID, newRecord)
		if err != nil {
			return fmt.Errorf("update record: %w", err)
		}
	}

	return nil
}

func getCurrentRecord(
	ctx context.Context,
	api *cloudflare.API,
	zoneID, recordName string,
) (*cloudflare.DNSRecord, error) {
	searchRecord := cloudflare.DNSRecord{
		Type: "A",
		Name: recordName,
	}
	records, err := api.DNSRecords(ctx, zoneID, searchRecord)
	if err != nil {
		return nil, fmt.Errorf("getting records: %w", err)
	}
	if len(records) == 0 {
		return nil, nil
	}
	return &records[0], nil
}
