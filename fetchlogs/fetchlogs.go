package fetchlogs

import (
	"time"
	"net/http"
	"context"
	"fmt"
	"io"
	"github.com/google/certificate-transparency-go/loglist3"
)

func Run() (*loglist3.LogList, error) {

	client := &http.Client{Timeout: 10 * time.Second}
	logListURL := "https://www.gstatic.com/ct/log_list/v3/log_list.json"

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, logListURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Failed to fetch log list from %s: %v\n", logListURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Received non-200 HTTP status code: %d %s\n", resp.StatusCode, resp.Status)
	}

	logListData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Failed to read response body: %v\n", err)
	}

	logList, err := loglist3.NewFromJSON(logListData)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse CT log list JSON: %v\n", err)
	}

	fmt.Printf("INFO: Fetched %d operators\n", len(logList.Operators))

	return logList, nil
}