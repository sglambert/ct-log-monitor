package similarity

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/agnivade/levenshtein"
	"github.com/xrash/smetrics"
	"golang.org/x/text/secure/precis"
)

type DomainEntry struct {
	Brand   string   `json:"brand"`
	Domain  string   `json:"domain"`
	Aliases []string `json:"aliases"`
}

type MatchResult struct {
	InputDomain      string    `json:"input_domain"`
	ConfigValue      string    `json:"matched_value"`
	MatchedBrand     string    `json:"matched_brand"`
	NumChecksMatched int       `json:"num_checks_matched"`
	ChecksMatched    []string  `json:"checks_matched"`
	Timestamp        time.Time `json:"timestamp"`
}

var configEntries []DomainEntry

func init() {
	err := loadConfig()
	if err != nil {
		fmt.Printf("ERROR loading config/monitor_domains.json: %v\n", err)
	}
}

func loadConfig() error {
	configPath := filepath.Join("config", "monitor_domains.json")
	f, err := os.Open(configPath)
	if err != nil {
		return err
	}
	defer f.Close()

	decoder := json.NewDecoder(f)
	return decoder.Decode(&configEntries)
}

func replaceHomoglyphs(s string) string {
    homoglyphs := map[rune]rune{
        'а': 'a', // Cyrillic Small Letter A (U+0430)
        'е': 'e', // Cyrillic Small Letter IE (U+0435) - Similar to Latin 'e'
        'ё': 'e', // Cyrillic Small Letter YO (U+0451) - Similar to 'e'
        'і': 'i', // Cyrillic Small Letter Dotted I (U+0456)
        'о': 'o', // Cyrillic Small Letter O (U+043E)
        'р': 'p', // Cyrillic Small Letter ER (U+0440)
        'ѕ': 's', // Cyrillic Small Letter ES (U+0455) / Greek Sigma (U+03C2)
        'ѵ': 'v', // Cyrillic Small Letter IZHITSA (U+0475) - Similar to 'v'
        'ԝ': 'w', // Cyrillic Small Letter WE (U+0449) - Similar to 'w'
        'ⅿ': 'm', // Fraktur Small M (U+212F) - Similar to 'm'
        'ⅼ': 'l', // Roman Numeral One (U+216C) - Similar to 'l'
        '𝖾': 'e', // Mathematical Sans-Serif Small E (U+1D586)
        '𝖺': 'a', // Mathematical Sans-Serif Small A (U+1D584)
        '𝖿': 'f', // Mathematical Sans-Serif Small F (U+1D588)
        'т': 't', // Cyrillic Small Letter TE (U+0442) -
        'ӏ': 'l', // Cyrillic Letter Palochka (U+04CF) -
        'с': 'c', // Cyrillic Small Letter ES (U+0441) - Similar to 'c'
        'м': 'm', // Cyrillic Small Letter EM (U+043C) - Similar to 'm'
    }

	var builder strings.Builder
	for _, r := range s {
		if replacement, ok := homoglyphs[r]; ok {
			builder.WriteRune(replacement)
		} else {
			builder.WriteRune(r)
		}
	}
	return builder.String()
}


func Run(inputDomains []string) {
	profile := precis.UsernameCaseMapped

	var allMatches []MatchResult

	for _, inputDomain := range inputDomains {
		normInput, err := profile.String(inputDomain)
		if err != nil {
			fmt.Printf("Failed to normalize input '%s': %v\n", inputDomain, err)
			continue
		}
		normInput = replaceHomoglyphs(normInput)

		var matchesForCurrentInput []MatchResult

		for _, entry := range configEntries {
			toCompare := append([]string{entry.Domain, entry.Brand}, entry.Aliases...)

			for _, target := range toCompare {
				normTarget, err := profile.String(target)
				if err != nil {
					fmt.Printf("Failed to normalize '%s': %v\n", target, err)
					continue
				}
				normTarget = replaceHomoglyphs(normTarget)

				var matchedChecks []string

				if dist := levenshtein.ComputeDistance(normInput, normTarget); dist <= 2 {
					matchedChecks = append(matchedChecks, "Levenshtein")
				}
				if sim := smetrics.JaroWinkler(normInput, normTarget, 0.7, 4); sim > 0.85 {
					matchedChecks = append(matchedChecks, "Jaro-Winkler")
				}
				if strings.Contains(normInput, normTarget) || strings.Contains(normTarget, normInput) {
					matchedChecks = append(matchedChecks, "Substring")
				}

				if len(matchedChecks) >= 2 {
					matchesForCurrentInput = append(matchesForCurrentInput, MatchResult{
						InputDomain:      inputDomain,
						ConfigValue:      target,
						MatchedBrand:     entry.Brand,
						NumChecksMatched: len(matchedChecks),
						ChecksMatched:    matchedChecks,
						Timestamp:        time.Now(),
					})
				}
			}
		}
		allMatches = append(allMatches, matchesForCurrentInput...)
	}

	if len(allMatches) > 0 {
		outputPath := filepath.Join("config", "logged_domains.json")

		var existing []MatchResult

		if existingFile, err := os.ReadFile(outputPath); err == nil {
			_ = json.Unmarshal(existingFile, &existing)
		}

		existing = append(existing, allMatches...)

		file, err := os.Create(outputPath)
		if err != nil {
			fmt.Printf("Failed to create output file: %v\n", err)
			return
		}
		defer file.Close()

		encoder := json.NewEncoder(file)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(existing); err != nil {
			fmt.Printf("Failed to write JSON: %v\n", err)
		} else {
			fmt.Printf("Similarity matches appended to %s\n", outputPath)
		}
	}
}