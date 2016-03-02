package main

import (
	"flag"
	"log"
	"syscall"

	"github.com/sevlyar/go-daemon"
)

var (
	daemonMode = flag.Bool("d", false, "run in daemon mode")
)

func main() {
	flag.Parse()

	context := &daemon.Context{
		PidFileName: "pid",
		PidFilePerm: 0644,
		LogFileName: "log",
		LogFilePerm: 0640,
		WorkDir:     "./",
		Umask:       027,
	}

	if *daemonMode {
		d, err := context.Search()
		if d == nil || err != nil {
			// no daemon process running, start it
			child, err := context.Reborn()
			if err != nil {
				log.Fatalln(err)
			}
			if child != nil {
				return
			}
			defer context.Release()
		} else {
			// check if the process is alive by sending signal 0
			err := d.Signal(syscall.Signal(0))
			if err == nil {
				// the process is running, don't do anything
				return
			} else {
				// no daemon process running, start it
				child, err := context.Reborn()
				if err != nil {
					log.Fatalln(err)
				}
				if child != nil {
					return
				}
				defer context.Release()
			}
		}
	}

	startIsmonitor(*daemonMode)
}

// func reloadHandler(sig os.Signal) error {
// 	log.Println("configuration reloaded")
// 	return nil
// }
