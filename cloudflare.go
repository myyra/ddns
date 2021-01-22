package main

import (
	"fmt"

	"github.com/cloudflare/cloudflare-go"
	"github.com/rs/zerolog"
)

func updateRecord(logger zerolog.Logger, api *cloudflare.API, zoneName string, recordName, ip string) error {
	zoneID, err := api.ZoneIDByName(zoneName)
	if err != nil {
		return fmt.Errorf("finding zone by name %s: %w", zoneName, err)
	}
	searchRecord := cloudflare.DNSRecord{
		Type: "A",
		Name: recordName,
	}
	wantRecord := cloudflare.DNSRecord{
		Type:    "A",
		Name:    recordName,
		TTL:     120,
		Content: ip,
	}

	records, err := api.DNSRecords(zoneID, searchRecord)
	if err != nil {
		return fmt.Errorf("getting records for zone %s: %w", zoneName, err)
	}
	if len(records) == 0 {
		logger.Info().Msg("record not found, creating")

		resp, err := api.CreateDNSRecord(zoneID, wantRecord)
		if err != nil {
			return fmt.Errorf("create record: %w", err)
		}
		if len(resp.Errors) != 0 {
			return fmt.Errorf("create record returned errors: %v", resp.Errors)
		}

		return nil
	} else {
		record := records[0]
		if record.Content != ip {
			logger.Info().Str("old_ip", record.Content).Msg("IP has changed, updating")

			err := api.UpdateDNSRecord(zoneID, records[0].ID, wantRecord)
			if err != nil {
				return fmt.Errorf("update record: %w", err)
			}
		}
	}

	return nil
}
