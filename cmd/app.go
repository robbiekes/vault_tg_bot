package main

import (
	"fmt"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	tele "gopkg.in/telebot.v4"
	"log"
	"os"
	"time"
	"vault_tg_bot/internal/handlers"
)

func appRun() {
	err := godotenv.Load()
	if err != nil {
		logrus.Fatal("error loading .env file")
	}

	tgBotToken := os.Getenv("TG_BOT_TOKEN")

	pref := tele.Settings{
		Token:  tgBotToken,
		Poller: &tele.LongPoller{Timeout: 10 * time.Second},
	}

	logrus.Info("Initializing new bot...")
	b, err := tele.NewBot(pref)
	if err != nil {
		log.Fatalf("error creating new bot: %s\n", err.Error())
		return
	}

	redisHost := os.Getenv("REDIS_HOST")
	redisPort := os.Getenv("REDIS_PORT")
	redisPwd := os.Getenv("REDIS_PWD")
	addr := fmt.Sprintf("%s:%s", redisHost, redisPort)

	logrus.Info("Initializing Redis...")
	r := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: redisPwd,
		DB:       0,
	})

	handlers.RunHandlers(b, r)
	b.Start()
}
