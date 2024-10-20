package main

import (
	"fmt"
	_ "github.com/joho/godotenv/autoload"
	"github.com/robfig/cron/v3"
	"log"
	"time"
	"yeetfile/backend/config"
	"yeetfile/backend/db"
	"yeetfile/backend/server"
	"yeetfile/backend/utils"
	"yeetfile/shared/constants"
)

func main() {
	defer db.Close()

	c := cron.New()
	var expiryCronID cron.EntryID
	var memberCronID cron.EntryID
	var limiterCronID cron.EntryID
	var bandwidthCronID cron.EntryID
	var downloadsCronID cron.EntryID

	var err error
	if config.IsDebugMode {
		expiryCronID, err = c.AddFunc("@every 1s", db.CheckExpiry)
	} else {
		go db.CheckExpiry()
		expiryCronID, err = c.AddFunc("@every 30s", db.CheckExpiry)
	}

	if err != nil {
		panic(err)
	}

	log.Println("Expiry cron task added!")

	if config.YeetFileConfig.BillingEnabled {
		// Enable membership inspection if billing is enabled
		go db.CheckMemberships()
		memberCronID, err = c.AddFunc("@daily", db.CheckMemberships)
		if err != nil {
			panic(err)
		}

		log.Println("Membership cron task added!")
	}

	go db.CheckBandwidth()
	bandwidthDuration := fmt.Sprintf("@every %dh", constants.BandwidthMonitorDuration*24)
	bandwidthCronID, err = c.AddFunc(bandwidthDuration, db.CheckBandwidth)
	if err != nil {
		panic(err)
	}

	log.Println("Bandwidth cron task added!")

	limiterCronID, err = c.AddFunc(
		fmt.Sprintf("@every %ds", constants.LimiterSeconds),
		server.ManageLimiters)
	if err != nil {
		panic(err)
	} else {
		log.Println("Limiter cron task added!")
	}

	downloadsCronID, err = c.AddFunc("@every 1h", db.CleanUpDownloads)
	if err != nil {
		panic(err)
	} else {
		log.Println("Download cleanup cron task added!")
	}

	if len(c.Entries()) > 0 && config.IsDebugMode {
		_, _ = c.AddFunc("@every 1m", func() {
			log.Println("~~ CRON MONITOR ~~")
			for _, e := range c.Entries() {
				if e.ID == expiryCronID {
					log.Println("Expiry | next run: " +
						e.Next.Format(time.RFC1123))
				} else if e.ID == memberCronID {
					log.Println("Memberships | next run: " +
						e.Next.Format(time.RFC1123))
				} else if e.ID == limiterCronID {
					log.Println("Limiter middleware | next run: " +
						e.Next.Format(time.RFC1123))
				} else if e.ID == downloadsCronID {
					log.Println("Downloads cleanup | next run: " +
						e.Next.Format(time.RFC1123))
				} else if e.ID == bandwidthCronID {
					log.Println("Bandwidth monitor | next run: " +
						e.Next.Format(time.RFC1123))
				}
			}
		})
	}

	c.Start()

	host := utils.GetEnvVar("YEETFILE_HOST", "localhost")
	port := utils.GetEnvVar("YEETFILE_PORT", "8090")

	server.Run(host, port)
}
