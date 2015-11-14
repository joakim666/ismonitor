package main

import (
    "fmt"
	"log"
	"os/exec"
	"io/ioutil"
	"encoding/json"
	"strings"
	"sort"
	"strconv"
)

type Config struct {
	DockerContainers 			[]string	`json:"docker_containers"`
	DiskUsagePercentWarning 	int 		`json:"disk_usage_percent_warning"`
	UptimeLoad5MinutesWarning	float64		`json:"uptime_load_5_minutes_warning"`
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


	for _, e := range errors {
		fmt.Print(e)
	}
}

func test_main() {
	configFile, err := ioutil.ReadFile("test/config.json")
	if err != nil {
		log.Fatal(err)
	}

	var config Config
	err = json.Unmarshal(configFile, &config)
	if err != nil {
		log.Fatal(err)
	}

	// 1. Verify running docker containers
	// sudo docker inspect --format='{{.Name}}' $(sudo docker ps -q --no-trunc)
	o, err := ioutil.ReadFile("test/output_docker.txt")
	if err != nil {
		log.Fatal(err)
	}
	var errors []string
	runningDockerErrors := verifyRunningDockerContainers(string(o), config.DockerContainers)
	errors = append(errors, runningDockerErrors...)

	// 2. Verify free space
	// df --output='source,pcent,target'
	o2, err := ioutil.ReadFile("test/output_df.txt")
	if err != nil {
		log.Fatal(err)
	}
	freeSpaceErrors := verifyFreeSpace(string(o2), config.DiskUsagePercentWarning)
	errors = append(errors, freeSpaceErrors...)

	// 3. Verify uptime load
	// uptime
	o3, err := ioutil.ReadFile("test/output_uptime.txt")
	if err != nil {
		log.Fatal(err)
	}
	uptimeErrors := verifyUptimeLoad(string(o3), config.UptimeLoad5MinutesWarning)
	errors = append(errors, uptimeErrors...)

	for _, e := range errors {
		fmt.Print(e)
	}
}

func verifyRunningDockerContainers(output string, expectedContainers []string) []string {
	var errors []string

	lines := strings.Split(output, "\n")

	// remove first char of container name from docker inspect output which is '/'
	for i, _ := range lines {
		if len(lines[i]) > 0 {
			lines[i] = lines[i][1:]
		}
	}

	sort.Strings(lines)

	for _, name := range expectedContainers {
		i := sort.Search(len(lines),
			func(i int) bool { return lines[i] >= name })
		if (i >= len(lines) || (i < len(lines) && lines[i] != name)) {
			errors = append(errors, fmt.Sprintf("Docker container '%s' is not running\n", name))
		}
	}

	return errors
}

func verifyFreeSpace(output string, diskUsagePercentWarning int) []string {
	var errors []string

	lines := strings.Split(output, "\n")

	for _, line := range lines[1:] { // skip header row
		columns := strings.Fields(line)

		if (len(columns) > 1) {
			//fmt.Println(columns)
			percentValue, err := strconv.Atoi(strings.Trim(columns[1], "%"))
			if err != nil {
				errors = append(errors, fmt.Sprint(err))
			} else {
				if (percentValue >= diskUsagePercentWarning) {
					errors = append(errors, fmt.Sprintf("Disk usage at %d percent\n", percentValue))
				}
			}
		}
	}

	return errors
}

func verifyUptimeLoad(output string, uptimeLoad5MinutesWarning float64) []string {
	var errors []string

	columns := strings.Fields(output)

	if (len(columns) >= 11) {
		floatValue, err := strconv.ParseFloat(strings.Trim(columns[11], ","), 64)
		if err != nil {
			errors = append(errors, fmt.Sprint(err))
		} else {
			if (floatValue >= uptimeLoad5MinutesWarning) {
				errors = append(errors, fmt.Sprintf("High load warning: %s\n", output))
			}
		}
	}

	return errors
}
