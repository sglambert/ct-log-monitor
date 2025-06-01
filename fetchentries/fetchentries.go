package fetchentries

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"

	"github.com/sglambert/ct-log-monitor/models"
	"github.com/sglambert/ct-log-monitor/processor"
	"github.com/sglambert/ct-log-monitor/similarity"
)

func FetchUpdatedEntries() error {
	file, err := os.Open("config/logbook.json")
	if err != nil {
		return fmt.Errorf("failed to open logbook.json: %w", err)
	}
	defer file.Close()

	var entries []models.LogbookEntry
	if err := json.NewDecoder(file).Decode(&entries); err != nil {
		return fmt.Errorf("failed to parse logbook.json: %w", err)
	}

	var wg sync.WaitGroup
	client := &http.Client{}

	for _, entry := range entries {
		cur := entry.STH.TreeSize
		prev := entry.STH.PreviousTreeSize

		switch {
		case cur < prev:
			fmt.Printf("WARNING: tree_size for '%s' shrank from %d to %d\n", entry.Name, prev, cur)
		case cur == prev:
			fmt.Printf("INFO: No new entries for '%s' (tree_size unchanged at %d)\n", entry.Name, cur)
		case cur > prev:
			fmt.Printf("INFO: New entries for '%s' from %d to %d\n", entry.Name, prev, cur-1)

			wg.Add(1)
			go func(entry models.LogbookEntry, start, end uint64) {
				defer wg.Done()
				logEntries, err := fetchEntries(client, entry, start, end)
				if err != nil {
					fmt.Printf("Error during fetchEntries: %v\n", err)
					return
				}

				for _, logEntry := range logEntries {
					certs, err := processor.ParseLogEntry(logEntry)
					if err != nil {
						fmt.Printf("Error parsing log entry: %v\n", err)
						continue
					}

					for _, cert := range certs {
						domainsToCheck := []string{}

						if cert.Subject.CommonName != "" {
							domainsToCheck = append(domainsToCheck, cert.Subject.CommonName)
						}

						if len(cert.DNSNames) > 0 {
							domainsToCheck = append(domainsToCheck, cert.DNSNames...)
						}

						if len(domainsToCheck) > 0 {
							similarity.Run(domainsToCheck)
						} else {
							fmt.Printf("No Common Name or SAN entries found for certificate within entry from '%s'\n", entry.Name)
						}
					}
				}
			}(entry, prev, cur-1)
		}
	}

	wg.Wait()
	fmt.Println("Done fetching updated entries.")
	return nil
}

func fetchEntries(client *http.Client, entry models.LogbookEntry, start, end uint64) ([]models.LogEntry, error) {
	entriesURL := fmt.Sprintf("%sct/v1/get-entries?start=%d&end=%d", entry.URL, start, end)
	req, err := http.NewRequest("GET", entriesURL, nil)
	if err != nil {
		return nil, fmt.Errorf("Failed to create request for '%s': %v\n", entry.Name, err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Failed to fetch entries for '%s': %v\n", entry.Name, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Non-200 response for '%s': %s\n", entry.Name, resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Failed to read entries response for '%s': %v\n", entry.Name, err)
	}

	var response models.CTLogEntriesResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("Failed to parse JSON for '%s': %w", entry.Name, err)
	}

	fmt.Printf("Received entries from '%s' (STH: %d to STH: %d), size: %d bytes\n", entry.Name, start, end, len(body))

	return response.Entries, nil
}