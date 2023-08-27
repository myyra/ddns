package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/rs/zerolog"
)

func getIP(ctx context.Context) (string, error) {
	log := zerolog.Ctx(ctx)

	log.Debug().Msg("getting IP from Cloudflare")
	url := "https://1.1.1.1/cdn-cgi/trace"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("constructing request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("doing request: %w", err)
	}
	defer resp.Body.Close()
	log.Debug().Msg("got response from Cloudflare")

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("reading response: %w", err)
	}

	values := make(map[string]string)
	kvpairs := strings.Split(string(data), "\n")
	if len(kvpairs) <= 1 {
		return "", fmt.Errorf("data doesn't look like k/v pairs, got: %v", kvpairs)
	}
	for _, kvpair := range kvpairs {
		log.Debug().Str("kvpair", kvpair).Msg("parsing k/v pair")
		if strings.Contains(kvpair, "=") {
			kv := strings.Split(kvpair, "=")
			if len(kv) != 2 {
				return "", fmt.Errorf("unable to get key/value from %v", kv)
			}
			values[kv[0]] = kv[1]
		}
	}

	return values["ip"], nil
}
