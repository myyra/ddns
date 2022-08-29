package main

import (
	"context"
	"fmt"

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
	zoneID, err := api.ZoneIDByName(zoneName)
	if err != nil {
		return fmt.Errorf("finding zone by name %s: %w", zoneName, err)
	}
	searchRecord := cloudflare.DNSRecord{
		Type: "A",
		Name: recordName,
	}
	newRecord := cloudflare.DNSRecord{
		Type:    "A",
		Name:    recordName,
		TTL:     120,
		Content: ip,
	}

	records, err := api.DNSRecords(ctx, zoneID, searchRecord)
	if err != nil {
		return fmt.Errorf("getting records for zone %s: %w", zoneName, err)
	}
	if len(records) == 0 {
		log.Info().Msg("record not found, creating")

		resp, err := api.CreateDNSRecord(ctx, zoneID, newRecord)
		if err != nil {
			return fmt.Errorf("create record: %w", err)
		}
		if len(resp.Errors) != 0 {
			return fmt.Errorf("create record returned errors: %v", resp.Errors)
		}
	} else {
		record := records[0]
		if record.Content != ip {
			log.Info().Str("old_ip", record.Content).Msg("IP has changed, updating")

			err := api.UpdateDNSRecord(ctx, zoneID, records[0].ID, newRecord)
			if err != nil {
				return fmt.Errorf("update record: %w", err)
			}
		}
	}

	return nil
}
