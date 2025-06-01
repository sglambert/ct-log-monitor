package main

import (
	"fmt"
	"github.com/sglambert/ct-log-monitor/fetchentries"
	"github.com/sglambert/ct-log-monitor/fetchlogs"
	"github.com/sglambert/ct-log-monitor/persistence"
)

func main() {

	loglist, err := fetchlogs.Run()
	if err != nil {
		fmt.Printf("ERROR: Error in fetchlogs during Run(): %v\n", err)
		return
	}

	err = persistence.Run(loglist)
	if err != nil {
		fmt.Printf("ERROR: Error in persistence during Run(): %v\n", err)
		return
	}

	err = persistence.FetchAndUpdateSTHs()
	if err != nil {
		fmt.Printf("ERROR: Error in persistence updating STHs: %v\n", err)
		return
	}

	err = fetchentries.FetchUpdatedEntries()
	if err != nil {
		fmt.Printf("ERROR: Error in fetch entries during update: %v\n", err)
		return
	}
}
