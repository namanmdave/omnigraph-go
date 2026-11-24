package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	omnigraph "github.com/namanmdave/omnigraph-go"
)

func main() {
	version := flag.String("version", omnigraph.ServerVersion, "omnigraph-server release version")
	outPath := flag.String("out", "spec/openapi.json", "output path for openapi.json")
	flag.Parse()

	url := fmt.Sprintf("https://raw.githubusercontent.com/ModernRelay/omnigraph/v%s/crates/omnigraph-server/openapi.json", *version)
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "fetch spec: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Fprintf(os.Stderr, "fetch %s: status %d\n", url, resp.StatusCode)
		os.Exit(1)
	}

	raw, err := io.ReadAll(io.LimitReader(resp.Body, 32<<20))
	if err != nil {
		fmt.Fprintf(os.Stderr, "read spec: %v\n", err)
		os.Exit(1)
	}
	if !json.Valid(raw) {
		fmt.Fprintln(os.Stderr, "downloaded spec is not valid JSON")
		os.Exit(1)
	}

	if err := os.WriteFile(*outPath, raw, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "write %s: %v\n", *outPath, err)
		os.Exit(1)
	}
	fmt.Printf("synced %s from v%s (%d bytes)\n", *outPath, *version, len(raw))
}
