package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/smtp"
	"os"
	"os/exec"
	"time"

	"github.com/robfig/cron"
)

type verificationError struct {
	title   string
	message string
}

type config struct {
	CronSchedule              *string            `json:"cron_schedule"`
	SMTP                      *smtpConfiguration `json:"smtp"`
	DockerContainers          []string           `json:"docker_containers"`
	DiskUsagePercentWarning   int                `json:"disk_usage_percent_warning"`
	UptimeLoad5MinutesWarning float64            `json:"uptime_load_5_minutes_warning"`
	ElkConfiguration          []elkConfiguration `json:"elk"`
}

type smtpConfiguration struct {
	Host string    `json:"host"`
	Port int       `json:"port"`
	Auth *smtpAuth `json:"auth"`
	From string    `json:"from"`
	To   []string  `json:"to"`
}

type smtpAuth struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type monitorJob struct {
	config *config
}

func (t monitorJob) Run() {
	//log.Println("Doing scheduled execution")
	runIsmonitor(*t.config)
}

func startIsmonitor(daemonMode bool) {
	configFile, err := ioutil.ReadFile("config.json")
	if err != nil {
		log.Fatalln(err)
	}

	var config config
	err = json.Unmarshal(configFile, &config)
	if err != nil {
		log.Fatalln(err)
	}

	if daemonMode && config.CronSchedule == nil {
		fmt.Println("Daemon mode but no cron schedule specified. Quitting.")
		os.Exit(1)
	}
	if !daemonMode && config.CronSchedule != nil {
		fmt.Println("Daemon mode not specified but a cron schedule specified. Quitting.")
		os.Exit(1)
	}

	if config.CronSchedule != nil {
		cron := cron.New()
		cron.AddJob(*config.CronSchedule, monitorJob{&config})
		cron.Start()
		defer cron.Stop()
		select {}
	} else {
		runIsmonitor(config)
	}
}

func runIsmonitor(config config) {
	var errors []verificationError

	// 1. Verify running docker containers
	// $ sudo docker inspect --format='{{.Name}}' $(sudo docker ps -q --no-trunc)
	o, err := exec.Command("bash", "-c", "docker inspect --format='{{.Name}}' $(docker ps -q --no-trunc)").Output()
	if err != nil {
		e := verificationError{title: "Docker verification error", message: fmt.Sprintf("Failed to run docker command: %s\n", fmt.Sprint(err))}
		errors = append(errors, e)
	}
	runningDockerErrors := verifyRunningDockerContainers(string(o), config.DockerContainers)
	errors = append(errors, runningDockerErrors...)

	// 2. Verify free space
	// df --output='source,pcent,target'
	o2, err := exec.Command("bash", "-c", "df --output='source,pcent,target'").Output() // need to run through bash to work on linux for some reason
	if err != nil {
		e := verificationError{title: "Disk usage verification error", message: fmt.Sprintf("Failed to run df command: %s\n", fmt.Sprint(err))}
		errors = append(errors, e)
	}
	freeSpaceErrors := verifyFreeSpace(string(o2), config.DiskUsagePercentWarning)
	errors = append(errors, freeSpaceErrors...)

	// 3. Verify load average
	o3, err := ioutil.ReadFile("/proc/loadavg")
	if err != nil {
		e := verificationError{title: "Load average verification error", message: fmt.Sprintf("Failed to read /proc/loadavg: %s\n", fmt.Sprint(err))}
		errors = append(errors, e)
	}
	uptimeErrors := verifyLoadAvg(string(o3), config.UptimeLoad5MinutesWarning)
	errors = append(errors, uptimeErrors...)

	elkErrors := doElkVerifications(config)
	errors = append(errors, elkErrors...)

	// report errors if any
	if len(errors) > 0 {
		err = report(config.SMTP, errors)
		if err != nil {
			log.Printf("Failed to report errors: %s\n", fmt.Sprint(err))
		}
	}
}

func report(smtpConfig *smtpConfiguration, errors []verificationError) error {
	var err error
	if smtpConfig == nil {
		// report to console
		for _, e := range errors {
			fmt.Printf("%s\n   %s", e.title, e.message)
		}
	} else {
		err = sendEmail(smtp.SendMail, *smtpConfig, time.Now(), errors)
	}

	return err
}
