package main

import (
	"context"
	"fmt"

	"github.com/cloudflare/cloudflare-go"
	"github.com/rs/zerolog"
)

func update(
	ctx context.Context,
	api *cloudflare.API,
	zoneContainer *cloudflare.ResourceContainer,
	record cloudflare.DNSRecord,
	ip string,
) (cloudflare.DNSRecord, error) {
	log := zerolog.Ctx(ctx)

	updateRecordParams := cloudflare.UpdateDNSRecordParams{
		ID:      record.ID,
		Type:    record.Type,
		Name:    record.Name,
		TTL:     ttlSeconds,
		Content: ip,
	}

	log.Debug().Interface("params", updateRecordParams).Msg("updating record")
	record, err := api.UpdateDNSRecord(ctx, zoneContainer, updateRecordParams)
	if err != nil {
		return cloudflare.DNSRecord{}, fmt.Errorf("update record: %w", err)
	}
	log.Debug().Interface("record", record).Msg("record updated")

	return record, nil
}

func createRecord(
	ctx context.Context,
	api *cloudflare.API,
	zoneContainer *cloudflare.ResourceContainer,
	recordName, ip string,
) (*cloudflare.DNSRecord, error) {
	log := zerolog.Ctx(ctx)

	createRecordParams := cloudflare.CreateDNSRecordParams{
		Type:    "A",
		Name:    recordName,
		TTL:     ttlSeconds,
		Content: ip,
	}

	log.Debug().Interface("params", createRecordParams).Msg("creating record")
	record, err := api.CreateDNSRecord(ctx, zoneContainer, createRecordParams)
	if err != nil {
		return nil, fmt.Errorf("create record: %w", err)
	}
	log.Debug().Interface("record", record).Msg("record created")

	return &record, nil
}

func firstRecordByName(
	ctx context.Context,
	api *cloudflare.API,
	zone *cloudflare.ResourceContainer,
	recordName string,
) (*cloudflare.DNSRecord, error) {
	log := zerolog.Ctx(ctx)

	log.Debug().Str("recordName", recordName).Msg("finding record")

	params := cloudflare.ListDNSRecordsParams{
		Type: "A",
		Name: recordName,
	}

	log.Debug().Interface("params", params).Msg("listing records")
	records, _, err := api.ListDNSRecords(ctx, zone, params)
	if err != nil {
		return nil, fmt.Errorf("getting records: %w", err)
	}

	logRecords := zerolog.Arr()
	for _, record := range records {
		logRecords = logRecords.Interface(record)
	}
	log.Debug().Array("records", logRecords).Msg("got records")

	if len(records) == 0 {
		return nil, nil
	}
	return &records[0], nil
}

func getZone(api *cloudflare.API, zoneName string) (*cloudflare.ResourceContainer, error) {
	zoneID, err := api.ZoneIDByName(zoneName)
	if err != nil {
		return nil, fmt.Errorf("finding zone by name %s: %w", zoneName, err)
	}
	return cloudflare.ZoneIdentifier(zoneID), nil
}
