# pjek-monitor

A simple cronbased docker container, disk usage and load monitoring application that alerts through 
cron (i.e. mail to root)

## Verifications

### Docker containers

Given a list of docker containers verifies that all of them are running.

### Disk usage

Alerts if disk usages goes over a configured threshold.

### Load average

Alerts if the 5 minutes load average is over a configured threshold.


## Build instructions

<code>$ docker run --rm -v "$PWD":/usr/src/pjek-monitor -w /usr/src/pjek-monitor golang:1.5.1 go build -v</code>


## Running

The application assumes the **config.json** file is in the current directory when the application is started. I.e.
run it like:

<code>$ cd foo && ./pjek-monitor</code>

I run it from cron with the following cron line:

<code>*/5 *  *   *   *     cd /root/pjek-monitor && ./pjek-monitor</code>

## TODO

* Allow running in daemon mode
* Add verification that queries Elasticsearch/logstash for number of log lines with given pattern in time period

## License

The MIT License (MIT)


