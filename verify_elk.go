package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
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

type ElkUrlTemplateData struct {
	Host string
	Port string
	Date string
}

type ElkBodyTemplateData struct {
	Query   string
	Minutes string
}

func doElkVerification(config ElkConfiguration) []string {
	var errors []string

	// if multiple indexes that will result in multiple calls to logstash
	// i.e. the results might be a combination of a query against the pre-midnight index and the
	// post-midnight index (as logstash does index rotation at midnight utc)
	indexes := elkIndexToUse(time.Now().UTC(), config.Minutes)
	var urls, err = makeUrls(config.Host, config.Port, indexes)
	if err != nil {
		errors = append(errors, fmt.Sprintf("Failed to make urls: %s\n", fmt.Sprint(err)))
		return errors
	}

	var outputs []string
	for _, url := range urls {
		body, err := makeBody(config.Query, config.Minutes)
		if err != nil {
			errors = append(errors, fmt.Sprintf("Failed to make elk request body: %s\n", fmt.Sprint(err)))
			return errors
		}

		resp, err := http.Post(url, "application/json", strings.NewReader(body))
		if err != nil {
			errors = append(errors, fmt.Sprintf("Failed to make elk request: %s\n", fmt.Sprint(err)))
			return errors
		}
		defer resp.Body.Close()
		res, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			errors = append(errors, fmt.Sprintf("Failed to read response from elk: %s\n", fmt.Sprint(err)))
			return errors
		}

		outputs = append(outputs, string(res))
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

func makeUrls(host string, port string, indexes []string) ([]string, error) {
	const elkUrlTemplate = "http://{{.Host}}:{{.Port}}/logstash-{{.Date}}/logs/_search"

	tmpl, err := template.New("url").Parse(elkUrlTemplate)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Failed to parse elk template: %s\n", fmt.Sprint(err)))
	}

	var urls []string
	for _, index := range indexes {
		templateData := ElkUrlTemplateData{host, port, index}

		var b bytes.Buffer
		err = tmpl.Execute(&b, templateData)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("Failed to parse elk template: %s\n", fmt.Sprint(err)))
		}

		urls = append(urls, b.String())
	}

	return urls, nil
}

func makeBody(query string, minutes int) (string, error) {
	const elkBodyTemplate = `{
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
	tmpl, err := template.New("body").Parse(elkBodyTemplate)
	if err != nil {
		return "", errors.New(fmt.Sprintf("Failed to parse elk template: %s\n", fmt.Sprint(err)))
	}

	templateData := ElkBodyTemplateData{template.JSEscapeString(query), fmt.Sprintf("%d", minutes)}

	var b bytes.Buffer
	err = tmpl.Execute(&b, templateData)
	if err != nil {
		return "", errors.New(fmt.Sprintf("Failed to parse elk template: %s\n", fmt.Sprint(err)))
	}

	return b.String(), nil
}
