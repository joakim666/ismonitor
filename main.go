package main

import (
    "fmt"
	"log"
	"os/exec"
	"io/ioutil"
	"encoding/json"
)

type Config struct {
	DockerContainers 			[]string			`json:"docker_containers"`
	DiskUsagePercentWarning 	int 				`json:"disk_usage_percent_warning"`
	UptimeLoad5MinutesWarning	float64				`json:"uptime_load_5_minutes_warning"`
	ElkConfiguration			[]ElkConfiguration	`json:"elk"`
}

type ElkConfiguration struct {
	Query			string	`json:"query"`
	MatchesEqual 	*int	`json:"matchesEquals"`
	MatchesAtLeast	*int	`json:"matchesAtLeast"`
	Minutes			int		`json:"minutes"`
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
