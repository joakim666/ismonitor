package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"text/template"
	"time"
)

type elkConfiguration struct {
	Host                string `json:"host"`
	Port                string `json:"port"`
	Query               string `json:"query"`
	MatchesEqual        *int   `json:"matchesEquals"`
	MatchesAtLeast      *int   `json:"matchesAtLeast"`
	Minutes             int    `json:"minutes"`
	NotificationMessage string `json:"notification_message"`
}

type elkURLTemplateData struct {
	Host string
	Port string
	Date string
}

type elkBodyTemplateData struct {
	Query   string
	Minutes string
}

func doElkVerifications(config config) []verificationError {
	var errors []verificationError

	for _, c := range config.ElkConfiguration {
		errors = append(errors, doElkVerification(c)...)
	}

	return errors
}

func doElkVerification(config elkConfiguration) []verificationError {
	var errors []verificationError

	// if multiple indexes that will result in multiple calls to logstash
	// i.e. the results might be a combination of a query against the pre-midnight index and the
	// post-midnight index (as logstash does index rotation at midnight utc)
	indexes := elkIndexToUse(time.Now().UTC(), config.Minutes)
	var urls, err = makeUrls(config.Host, config.Port, indexes)
	if err != nil {
		e := verificationError{title: "Elk verification error", message: fmt.Sprintf("Failed to make urls: %s\n", fmt.Sprint(err))}
		errors = append(errors, e)
		return errors
	}

	var outputs []string
	for _, url := range urls {
		body, err := makeBody(config.Query, config.Minutes)
		if err != nil {
			e := verificationError{title: "Elk verification error", message: fmt.Sprintf("Failed to make elk request body: %s\n", fmt.Sprint(err))}
			errors = append(errors, e)
			return errors
		}

		resp, err := http.Post(url, "application/json", strings.NewReader(body))
		if err != nil {
			e := verificationError{title: "Elk verification error", message: fmt.Sprintf("Failed to make elk request: %s\n", fmt.Sprint(err))}
			errors = append(errors, e)
			return errors
		}
		defer resp.Body.Close()
		res, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			e := verificationError{title: "Elk verification error", message: fmt.Sprintf("Failed to read response from elk: %s\n", fmt.Sprint(err))}
			errors = append(errors, e)
			return errors
		}

		outputs = append(outputs, string(res))
	}

	if config.MatchesEqual != nil {
		e := verifyElkExpectedNoOfMatches(outputs, *config.MatchesEqual, config.NotificationMessage)
		errors = append(errors, e...)
	} else {
		e := verifyElkAtLeastNoOfMatches(outputs, *config.MatchesAtLeast, config.NotificationMessage)
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

func verifyElkExpectedNoOfMatches(outputs []string, expectedMatches int, notificationMessage string) []verificationError {
	var errors []verificationError

	var matches []ElkHit
	var total = 0

	// parse the json outputs. Collect the matches and sum the total number of matches
	for _, o := range outputs {
		var res ElkResult
		err := json.Unmarshal([]byte(o), &res)
		if err != nil {
			e := verificationError{title: "Elk verification error", message: fmt.Sprintf("Failed to parse json output file: %s\n", fmt.Sprint(err))}
			return append(errors, e)
		}
		matches = append(matches, res.Results.Hits...)
		total += res.Results.Total
	}

	if total != expectedMatches {
		if total == 0 {
			e := verificationError{title: notificationMessage, message: fmt.Sprintf("Expected %d matches but was 0\n", expectedMatches)}
			errors = append(errors, e)
		} else {
			for _, hit := range matches {
				//errors = append(errors, fmt.Sprintf("Expected %d matches but was %d:\n", expectedMatches, total))
				e := verificationError{title: notificationMessage, message: fmt.Sprintf("%s %s %s\n", hit.Source.Timestamp, hit.Source.DockerName, hit.Source.Message)}
				errors = append(errors, e)
			}
		}
	}

	return errors
}

func verifyElkAtLeastNoOfMatches(outputs []string, atleast int, notificationMessage string) []verificationError {
	var errors []verificationError

	var matches []ElkHit
	var total = 0

	// parse the json outputs. Collect the matches and sum the total number of matches
	for _, o := range outputs {
		var res ElkResult
		err := json.Unmarshal([]byte(o), &res)
		if err != nil {
			e := verificationError{title: notificationMessage, message: fmt.Sprintf("Failed to parse json output file: %s\n", fmt.Sprint(err))}
			errors = append(errors, e)
			return errors
		}
		matches = append(matches, res.Results.Hits...)
		total += res.Results.Total
	}

	if total < atleast {
		e := verificationError{title: notificationMessage, message: fmt.Sprintf("Expected at least %d matches but was %d\n", atleast, total)}
		errors = append(errors, e)
		if total > 0 {
			e := verificationError{
				title: "Elk verification error",
				message: fmt.Sprintf("One of the matching lines: %s %s %s\n",
					matches[0].Source.Timestamp,
					matches[0].Source.DockerName,
					matches[0].Source.Message)}
			errors = append(errors, e)
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
	const elkURLTemplate = "http://{{.Host}}:{{.Port}}/logstash-{{.Date}}/logs/_search"

	tmpl, err := template.New("url").Parse(elkURLTemplate)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse elk template: %s\n", fmt.Sprint(err))
	}

	var urls []string
	for _, index := range indexes {
		templateData := elkURLTemplateData{host, port, index}

		var b bytes.Buffer
		err = tmpl.Execute(&b, templateData)
		if err != nil {
			return nil, fmt.Errorf("Failed to parse elk template: %s\n", fmt.Sprint(err))
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
		return "", fmt.Errorf("Failed to parse elk template: %s\n", fmt.Sprint(err))
	}

	templateData := elkBodyTemplateData{template.JSEscapeString(query), fmt.Sprintf("%d", minutes)}

	var b bytes.Buffer
	err = tmpl.Execute(&b, templateData)
	if err != nil {
		return "", fmt.Errorf("Failed to parse elk template: %s\n", fmt.Sprint(err))
	}

	return b.String(), nil
}
