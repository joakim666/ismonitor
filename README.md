[![Build Status](https://travis-ci.org/joakim666/ismonitor.svg)](https://travis-ci.org/joakim666/ismonitor)

Version: 0.0.2

# ismonitor

A simple monitoring tool. It can run both as a daemon or from cron.

## Verifications

### Docker containers

Given a list of docker containers verifies that all of them are running.

### Disk usage

Alerts if disk usages goes over a configured threshold.

### Load average

Alerts if the 5 minutes load average is over a configured threshold.

### Assertions against logstash queries

E.g. verify no matches for the string 'ERROR' in all log files the last 5 minutes or that the string 'successful' 
appeared at least 3 times.


## Build instructions

<code>$ docker run --rm -v "$PWD":/go/src/ismonitor -w /go/src/ismonitor golang:1.6 bash -c 'go get && go build -v'</code>


## Running

The application assumes the presence of the **config.json** file in the current directory.

### Running from cron

For example run it from cron every five minutes:

<code>*/5 *  *   *   *     cd /root/ismonitor && ./ismonitor</code>

When executing from cron ismonitor can either do the error reporting to stdout (which cron will pick up and do what
it's been configured to do with it) or sent as email through the configured smtp server.

### Running as a daemon

Start it as a daemon using the **-d** flag. I.e.

<code>$ ./ismonitor -d</code>

It writes the pid in a file named **pid** and the log output in a file named **log** in the current directory.

If the application is started with the -d flag while the daemon process is running nothing will happen. I used this
to trigger restarts of the daemon process from cron in case it would quit, using a cron entry like this:

<code>0 *  *   *   *     cd /root/ismonitor && ./ismonitor -d</code>

I.e. every hour the ismonitor applicaton will be started unless it already is running.

When executing as a daemon the **cron_schedule** configuration is required in the config.json file.

## Error reporting

Error reporting can be done either by writing to stdout or by sending mail through an SMTP server.

If SMTP configuration is included in the configuration file error reporting will be done by sending email. Otherwise the error reporting it done to standard output. If ismonitor is executed in daemon mode it's standard output will be redirected to a file named **log**. If there is no SMTP configuration the error reporting will hence be found in the log file. 


## License

The MIT License (MIT)


