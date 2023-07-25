package main

import (
	"fmt"
	"github.com/go-co-op/gocron"
	"github.com/xacnio/backupper/internal/backup"
	"github.com/xacnio/backupper/internal/config"
	"github.com/xacnio/backupper/internal/utils"
	"github.com/xacnio/backupper/internal/utils/logger"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

const VERSION = "0.0.6"

func main() {
	// Load config and logger
	config.ReadConfig()
	logger.Init()

	// Load timezone
	if config.Get().Timezone != nil {
		utils.LoadLocation(*config.Get().Timezone)
	} else {
		utils.LoadLocation(os.Getenv("TZ"))
	}

	// Create scheduler and load all the backups
	s := gocron.NewScheduler(utils.TimeLocation)
	backups := utils.ConvertToStruct[[]backup.Backup](config.Get().Backups)
	for i, bup := range backups {
		var err error
		if strings.Count(bup.CronExpression, " ") == 5 {
			backups[i].Job, err = s.CronWithSeconds(bup.CronExpression).Do(bup.CreateFunc())
		} else {
			backups[i].Job, err = s.Cron(bup.CronExpression).Do(bup.CreateFunc())
		}
		if err != nil {
			logger.Main.Errorw("cron error", "name", bup.Name, "error", err)
		}
	}

	// Print start message
	go WaitBlockingAndPrint(s, &backups)

	// Detect interrupt signal
	sigc := GetSignalChannel()
	go DetectSignal(sigc, &backups)

	// Start all the pending jobs and run them forever
	s.StartBlocking()
}

func PrintStartMessage(version string, backups *[]backup.Backup) {
	fmt.Println("Backupper v" + version)
	fmt.Println("Total schedules: " + fmt.Sprintf("%d", len(*backups)))
	for _, b := range *backups {
		scheduled := b.Job.NextRun()
		fmt.Printf("  - %s - Next run: %s (%s)\n", b.Name, scheduled.Format("2006-01-02 15:04:05"), b.CronExpression)
	}
}

func WaitBlockingAndPrint(s *gocron.Scheduler, backups *[]backup.Backup) {
	for {
		if s.IsRunning() {
			PrintStartMessage(VERSION, backups)
			return
		} else {
			time.Sleep(1 * time.Second)
		}
	}
}

func GetSignalChannel() chan os.Signal {
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	return sigc
}

func DetectSignal(sigc chan os.Signal, backups *[]backup.Backup) {
	s := <-sigc
	fmt.Println("Received signal: " + s.String())
	waiting := false
	for {
		if !waiting {
			for _, b := range *backups {
				if b.Job.IsRunning() {
					fmt.Println("Waiting for backup to finish to exit...")
					waiting = true
				}
			}
		}
		if !waiting {
			os.Exit(0)
		} else {
			allFinished := true
			for _, b := range *backups {
				if b.Job.IsRunning() {
					allFinished = false
					break
				}
			}
			if allFinished {
				fmt.Println("All backups finished, exiting...")
				os.Exit(0)
			}
			time.Sleep(1 * time.Second)
			continue
		}
	}
}
