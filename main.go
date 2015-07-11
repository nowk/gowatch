package main

import (
	"flag"
	"log"
	"strings"
	"time"
)

var (
	wait           = flag.Duration("wait", time.Second, "# seconds to wait before restarting")
	ignore         = flag.String("ignore", "", "comma delimited paths to ignore in the file watcher")
	debug          = flag.Bool("debug", true, "enabled debug print statements")
	pwd            = flag.String("dir", ".", "working directory ")
	restartOnError = flag.Bool("onerror", true, "If a non-zero exit code should restart the lint/build/test/run process")
	//test           = flag.Bool("test", false, "run go test on reload")
	//lint           = flag.Bool("lint", false, "run go lint on reload")
	//stepUpdates = make(chan bool)
	ignorePaths = []string{}
)

func init() {
	flag.Parse()
}

func main() {
	ignorePaths = strings.Split(*ignore, ",")

	if *debug {
		log.Println("Debug mode enabled.")
		if !*restartOnError {
			log.Println("\tRestart on error disabled")
		}
		for _, files := range ignorePaths {
			log.Println("\tignoring", files)
		}
	}

	proj := createProject(*pwd)
	cwd := proj.WorkingDirectory()
	*pwd = cwd

	fileUpdates := getWatch(cwd)

	var lastErr error = nil
	for {

		select {
		case <-fileUpdates:
			if *debug {
				log.Println("File update, starting build steps.")
			}

			if proj.kill != nil {
				proj.kill <- true
			}

			lastErr = nil
			time.Sleep(*wait)
		default:
			if !*restartOnError && lastErr != nil {
				continue
			} else if proj.kill == nil {
				if lastErr != nil {
					log.Println(lastErr)
				}

				errorChan := proj.RunSteps()
				lastErr = <-errorChan

				if lastErr != nil {
					if *debug {
						log.Println("run step error")
					}
				}
				time.Sleep(*wait)
			}
		}

	}
}
