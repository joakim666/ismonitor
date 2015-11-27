package main
import (
	"encoding/json"
	"fmt"
	"time"
)

type ElkResult struct {
	Results		ElkHits	`json:"hits"`
}

type ElkHits struct {
	Total		int			`json:"total"`
	Hits		[]ElkHit	`json:"hits"`
}

type ElkHit struct {
	Source		ElkHitSource	`json:"_source"`
}

type ElkHitSource struct {
	Message		string			`json:"message"`
	DockerName	string			`json:"docker.name"`
	Timestamp	string			`json:"@timestamp"`
}

func verifyElkExpectedNoOfMatches(outputs []string, expectedMatches int) []string {
	var errors []string

	var matches []ElkHit
	var total = 0

	// parse the json outputs. Collect the matches and sum the total number of matches
	for _, o := range outputs {
		var res ElkResult
		err := json.Unmarshal([]byte(o), &res)
		if err != nil {
			errors = append(errors, fmt.Sprintf("Failed to parse json output file: %s\n", fmt.Sprint(err)))
			return errors
		}
		matches = append(matches, res.Results.Hits...)
		total += res.Results.Total
	}

	if total != expectedMatches {
		if total == 0 {
			errors = append(errors, fmt.Sprintf("Expected %d matches but was 0\n", expectedMatches))
		} else {
			for _, hit := range matches {
				//errors = append(errors, fmt.Sprintf("Expected %d matches but was %d:\n", expectedMatches, total))
				errors = append(errors, fmt.Sprintf("%s %s %s\n", hit.Source.Timestamp, hit.Source.DockerName, hit.Source.Message))
			}
		}
	}

	return errors
}

func verifyElkAtLeastNoOfMatches(outputs []string, atleast int) []string {
	var errors []string

	var matches []ElkHit
	var total = 0

	// parse the json outputs. Collect the matches and sum the total number of matches
	for _, o := range outputs {
		var res ElkResult
		err := json.Unmarshal([]byte(o), &res)
		if err != nil {
			errors = append(errors, fmt.Sprintf("Failed to parse json output file: %s\n", fmt.Sprint(err)))
			return errors
		}
		matches = append(matches, res.Results.Hits...)
		total += res.Results.Total
	}

	if total < atleast {
		errors = append(errors, fmt.Sprintf("Expected at least %d matches but was %d\n", atleast, total))
		if total > 0 {
			errors = append(errors, fmt.Sprintf("One of the matching lines: %s %s %s\n", matches[0].Source.Timestamp,
				matches[0].Source.DockerName, matches[0].Source.Message))
		}
	}

	return errors
}

func formatDateForElkIndex(time time.Time) string {
	return time.Format("2006.01.02")
}

func elkIndexToUse(time time.Time, minutes int) string {

//	time.

	return formatDateForElkIndex(time)
}
