package persistence

import (
	"github.com/sglambert/ct-log-monitor/models"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"github.com/google/certificate-transparency-go/loglist3"
)

type LogEntry struct {
	Name string      `json:"name"`
	URL  string      `json:"url"`
	STH  *models.STH `json:"sth,omitempty"`
}

func Run(data *loglist3.LogList) error {

	var entries []LogEntry

	existingEntries := make(map[string]LogEntry)
	filePath := filepath.Join("config", "logbook.json")

	if fileData, err := os.ReadFile(filePath); err == nil {
		var existing []LogEntry
		if err := json.Unmarshal(fileData, &existing); err == nil {
			for _, e := range existing {
				existingEntries[e.URL] = e
			}
			entries = existing
		}
	}

	activeLogs := 0
	newLogs := 0

	for _, op := range data.Operators {
		fmt.Printf("Operator: %s\n", op.Name)
		for _, log := range op.Logs {
			status := log.State.LogStatus()

			if status == loglist3.UsableLogStatus || status == loglist3.QualifiedLogStatus {
				activeLogs++
				if _, exists := existingEntries[log.URL]; !exists {
					newEntry := LogEntry{
						Name: log.Description,
						URL:  log.URL,
					}
					entries = append(entries, newEntry)
					existingEntries[log.URL] = newEntry
					newLogs++
				}
			}
		}
	}

	if err := os.MkdirAll("config", os.ModePerm); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "    ")
	if err := encoder.Encode(entries); err != nil {
		return fmt.Errorf("failed to write JSON: %w", err)
	}

	fmt.Printf("Total active logs found: %d\n", activeLogs)
	fmt.Printf("New logs added to logbook.json: %d\n", newLogs)

	return nil
}

func FetchAndUpdateSTHs() error {
	path := filepath.Join("config", "logbook.json")

	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read logbook.json: %w", err)
	}

	var entries []models.LogbookEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return fmt.Errorf("failed to unmarshal logbook.json: %w", err)
	}

	client := &http.Client{}
	for i := range entries {
		log := &entries[i]
		sthURL := fmt.Sprintf("%sct/v1/get-sth", log.URL)

		req, err := http.NewRequest("GET", sthURL, nil)
		if err != nil {
			fmt.Printf("Failed to create request for %s: %v\n", log.Name, err)
			continue
		}

		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("Failed to get STH for %s: %v\n", log.Name, err)
			continue
		}
		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			fmt.Printf("Non-200 response for %s: %s\n", log.Name, resp.Status)
			continue
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return fmt.Errorf("Failed to read body for %s: %v\n", log.Name, err)
		}

		var newSTH models.CTGetSTHResponse
		if err := json.Unmarshal(body, &newSTH); err != nil {
			return fmt.Errorf("Failed to parse STH JSON for %s: %v\n", log.Name, err)
		}

		log.STH.PreviousTreeSize = log.STH.TreeSize
		log.STH.PreviousTimestamp = log.STH.Timestamp

		log.STH.TreeSize = newSTH.TreeSize
		log.STH.Timestamp = newSTH.Timestamp

		fmt.Printf("Updated %s: tree_size %d -> %d\n", log.Name, log.STH.PreviousTreeSize, log.STH.TreeSize)
	}

	updatedJSON, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal updated logbook: %w", err)
	}

	tmpFile := filepath.Join(filepath.Dir(path), "logbook.tmp.json")
	if err := os.WriteFile(tmpFile, updatedJSON, 0644); err != nil {
		return fmt.Errorf("failed to write temp logbook file: %w", err)
	}

	if err := os.Rename(tmpFile, path); err != nil {
		return fmt.Errorf("failed to rename temp logbook file: %w", err)
	}

	fmt.Printf("Logbook updated successfully.")
	return nil
}
