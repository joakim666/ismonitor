package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"text/template"
	"time"
)

func doElkVerifications(config Config) []string {
	var errors []string

	for _, c := range config.ElkConfiguration {
		errors = append(errors, doElkVerification(c)...)
	}

	return errors
}

type ElkTemplateData struct {
	Date    string
	Query   string
	Minutes string
}

func doElkVerification(config ElkConfiguration) []string {
	const elkTemplate = `curl -XPOST localhost:9200/logstash-{{.Date}}/logs/_search -d '{
  "query": {
    "filtered": {
      "query": {
        "query_string": {
          "query": "{{.Query}}"
        }
      },
      "filter": {
        "bool": {
          "must": [
            {
              "range": {
                "@timestamp": {
                  "gte": "now-{{.Minutes}}m"
                }
              }
            }
          ],
          "must_not": []
        }
      }
    }
  },
  "size": 500,
  "sort": {
    "@timestamp": "desc"
  },
  "fields": [
    "_source"
  ],
  "script_fields": {},
  "fielddata_fields": [
    "timestamp",
    "@timestamp"
  ]
}'
`
	var errors []string

	tmpl, err := template.New("elk").Parse(elkTemplate)
	if err != nil {
		errors = append(errors, fmt.Sprintf("Failed to parse elk template: %s\n", fmt.Sprint(err)))
		return errors
	}

	indexes := elkIndexToUse(time.Now().UTC(), config.Minutes)

	// if multiple indexes that will result in multiple calls to logstash
	// i.e. the results might be a combination of a query against the pre-midnight index and the
	// post-midnight index (as logstash does index rotation at midnight utc)

	var outputs []string
	for _, index := range indexes {
		templateData := ElkTemplateData{index,
			template.JSEscapeString(config.Query), fmt.Sprintf("%d", config.Minutes)}

		var b bytes.Buffer
		err = tmpl.Execute(&b, templateData)
		if err != nil {
			errors = append(errors, fmt.Sprintf("Failed to parse elk template: %s\n", fmt.Sprint(err)))
			return errors
		}

		cmd := fmt.Sprintf("docker exec elk %s", b.String())
		o, err := exec.Command("bash", "-c", cmd).Output()
		if err != nil {
			errors = append(errors, fmt.Sprintf("Failed to run docker command: %s\n", fmt.Sprint(err)))
		}

		outputs = append(outputs, string(o))
	}

	if config.MatchesEqual != nil {
		e := verifyElkExpectedNoOfMatches(outputs, *config.MatchesEqual)
		errors = append(errors, e...)
	} else {
		e := verifyElkAtLeastNoOfMatches(outputs, *config.MatchesAtLeast)
		errors = append(errors, e...)
	}

	return errors
}

type ElkResult struct {
	Results ElkHits `json:"hits"`
}

type ElkHits struct {
	Total int      `json:"total"`
	Hits  []ElkHit `json:"hits"`
}

type ElkHit struct {
	Source ElkHitSource `json:"_source"`
}

type ElkHitSource struct {
	Message    string `json:"message"`
	DockerName string `json:"docker.name"`
	Timestamp  string `json:"@timestamp"`
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

// elkIndexToUse figures out what logstash index to use based on the current time and the time period for the
// logstash query. If the query spans midnight UTC, when logstash does daily index rotation, this function
// will return both the pre-midnight index and the post-midnight index
func elkIndexToUse(now time.Time, minutes int) []string {
	then := now.Add(time.Duration(-1*minutes) * time.Minute)

	if now.Day() != then.Day() {
		return []string{formatDateForElkIndex(then), formatDateForElkIndex(now)}
	} else {
		return []string{formatDateForElkIndex(now)}
	}

}
