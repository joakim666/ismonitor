package main

import (
    "fmt"
	"log"
	"os/exec"
	"io/ioutil"
	"encoding/json"
	"text/template"
	"bytes"
	"time"
)

type Config struct {
	DockerContainers 			[]string			`json:"docker_containers"`
	DiskUsagePercentWarning 	int 				`json:"disk_usage_percent_warning"`
	UptimeLoad5MinutesWarning	float64				`json:"uptime_load_5_minutes_warning"`
	ElkConfiguration			[]ElkConfiguration	`json:"elk"`
}

type ElkConfiguration struct {
	Query			string	`json:"query"`
	MatchesEqual 	*int		`json:"matchesEquals"`
	MatchesAtLeast	*int		`json:"matchesAtLeast"`
}


func main() {
	configFile, err := ioutil.ReadFile("config.json")
	if err != nil {
		log.Fatal(err)
	}

	var config Config
	err = json.Unmarshal(configFile, &config)
	if err != nil {
		log.Fatal(err)
	}

	var errors []string

	// 1. Verify running docker containers
	// $ sudo docker inspect --format='{{.Name}}' $(sudo docker ps -q --no-trunc)
	o, err := exec.Command("bash", "-c", "docker inspect --format='{{.Name}}' $(docker ps -q --no-trunc)").Output()
	if err != nil {
		errors = append(errors, fmt.Sprintf("Failed to run docker command: %s\n", fmt.Sprint(err)))
	}
	runningDockerErrors := verifyRunningDockerContainers(string(o), config.DockerContainers)
	errors = append(errors, runningDockerErrors...)

	// 2. Verify free space
	// df --output='source,pcent,target'
	o2, err := exec.Command("bash", "-c", "df --output='source,pcent,target'").Output() // need to run through bash to work on linux for some reason
	if err != nil {
		errors = append(errors, fmt.Sprintf("Failed to run df command: %s\n", fmt.Sprint(err)))
	}
	freeSpaceErrors := verifyFreeSpace(string(o2), config.DiskUsagePercentWarning)
	errors = append(errors, freeSpaceErrors...)

	// 3. Verify uptime load
	// uptime
	o3, err := exec.Command("uptime").Output()
	if err != nil {
		errors = append(errors, fmt.Sprintf("Failed to run uptime command: %s\n", fmt.Sprint(err)))
	}
	uptimeErrors := verifyUptimeLoad(string(o3), config.UptimeLoad5MinutesWarning)
	errors = append(errors, uptimeErrors...)


	elkErrors := doElkVerifications(config)
	errors = append(errors, elkErrors...)


	for _, e := range errors {
		fmt.Print(e)
	}
}

func doElkVerifications(config Config) []string {
	var errors []string

	for _, c := range(config.ElkConfiguration) {
		errors = append(errors, doElkVerification(c)...)
	}

	return errors
}

type ElkTemplateData struct {
	Date	string
	Query	string
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
                  "gte": "now-5m"
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

	templateData := ElkTemplateData{formatDateForElkIndex(time.Now()), template.JSEscapeString(config.Query)}

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

	if config.MatchesEqual != nil {
		e := verifyElkExpectedNoOfMatches(string(o), *config.MatchesEqual)
		errors = append(errors, e...)
	} else {
		e:= verifyElkAtLeastNoOfMatches(string(o), *config.MatchesAtLeast)
		errors = append(errors, e...)
	}

	return errors
}
