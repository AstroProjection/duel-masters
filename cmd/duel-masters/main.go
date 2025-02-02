package main

import (
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"time"

	"duel-masters/api"
	"duel-masters/db"
	"duel-masters/game"
	"duel-masters/game/cards"
	"duel-masters/game/match"

	"github.com/sirupsen/logrus"
)

func main() {

	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetLevel(logrus.DebugLevel)

	rand.Seed(time.Now().UnixNano())

	logrus.Info("Starting..")

	go checkForAutoRestart()

	for _, set := range cards.Sets {
		for uid, ctor := range *set {
			match.AddCard(uid, ctor)
		}
	}

	go game.GetLobby().StartTicker()

	api.CreateCardCache()

	db.Connect(os.Getenv("mongo_uri"), os.Getenv("mongo_name"))

	api.Start(os.Getenv("port"))

}

func checkForAutoRestart() {

	if os.Getenv("restart_after") == "" {
		logrus.Debug("No autorestart policy found")
		return
	}

	n, err := strconv.Atoi(os.Getenv("restart_after"))

	if err != nil {
		panic(err)
	}

	d := time.Now().Add(time.Second * time.Duration(n))

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	notified := false

	logrus.Info(fmt.Sprintf("Scheduled to shutdown %s", d.Format("2 Jan 2006 15:04")))

	for range ticker.C {

		if time.Now().After(d) {
			logrus.Info("Performing scheduled shutdown")
			os.Exit(0)
		}

		// less than 2 hours until restart and have not yet notified
		if time.Now().Add(2*time.Hour).After(d) && !notified {
			notified = true

			game.PinMessage(fmt.Sprintf("Scheduled restart in time:%v", d.Unix()))
		}

	}

}
