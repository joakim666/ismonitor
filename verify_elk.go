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

func verifyElkExpectedNoOfMatches(output string, expectedMatches int) []string {
	var errors []string

	var elkResult ElkResult
	err := json.Unmarshal([]byte(output), &elkResult)
	if err != nil {
		errors = append(errors, fmt.Sprintf("Failed to parse json output file: %s\n", fmt.Sprint(err)))
		return errors
	}

	if elkResult.Results.Total != expectedMatches {
		if elkResult.Results.Total == 0 {
			errors = append(errors, fmt.Sprintf("Expected %d matches but was 0\n", expectedMatches))
		} else {
			for _, hit := range elkResult.Results.Hits {
				errors = append(errors, fmt.Sprintf("%s %s %s\n", hit.Source.Timestamp, hit.Source.DockerName, hit.Source.Message))
			}
		}
	}

	return errors
}

func verifyElkAtLeastNoOfMatches(output string, atleast int) []string {
	var errors []string

	var elkResult ElkResult

	err := json.Unmarshal([]byte(output), &elkResult)
	if err != nil {
		errors = append(errors, fmt.Sprintf("Failed to parse json output file: %s\n", fmt.Sprint(err)))
		return errors
	}

	if elkResult.Results.Total < atleast {
		errors = append(errors, fmt.Sprintf("Expected at least %d matches but was %d\n", atleast, elkResult.Results.Total))
		if elkResult.Results.Total > 0 {
			errors = append(errors, fmt.Sprintf("One of the matching lines: %s %s %s\n", elkResult.Results.Hits[0].Source.Timestamp,
				elkResult.Results.Hits[0].Source.DockerName, elkResult.Results.Hits[0].Source.Message))
		}
	}

	return errors
}

func formatDateForElkIndex(time time.Time) string {
	return time.Format("2006.01.02")
}
