package daemon

import (
	"errors"
	"log"
	"net/http"
	"os"

	"github.com/Ars2014/IT-CatlangBot/dbHelper"
	"github.com/Ars2014/IT-CatlangBot/handlers"
	"github.com/Ars2014/IT-CatlangBot/webchart"
	"github.com/Syfaro/telegram-bot-api"
	"github.com/boltdb/bolt"
)

var db *bolt.DB

func NewBot() (*tgbotapi.BotAPI, error) {
	token := os.Getenv("token")
	debug := os.Getenv("debug")
	if token == "" {
		return nil, errors.New("no 'token' were specified in .env file")
	}
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}
	if debug == "true" {
		bot.Debug = true
	}
	return bot, nil
}

type ChartHandler struct {
}

func (*ChartHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	tmpl, err := webchart.ChartPage(db)
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte("Произошла ошибка!"))
		panic(err)
	}
	w.Write([]byte(tmpl))
}

func Run() error {
	bot, err := NewBot()
	if err != nil {
		return err
	}

	db, err = dbHelper.SetupDB()
	if err != nil {
		return err
	}

	log.Printf("[Daemon] Authorized on account %s", bot.Self.UserName)

	var updates tgbotapi.UpdatesChannel

	if os.Getenv("dev") == "true" {
		u := tgbotapi.NewUpdate(0)
		u.Timeout = 60

		updates, err = bot.GetUpdatesChan(u)

		if err != nil {
			return err
		}
	} else {
		_, err = bot.SetWebhook(tgbotapi.NewWebhook(os.Getenv("base_url") + os.Getenv("webhook_url")))
		if err != nil {
			return err
		}

		updates = bot.ListenForWebhook("/")
		go http.ListenAndServe(":"+os.Getenv("webhook_server_port"), nil)

	}

	go http.ListenAndServe(":"+os.Getenv("chart_server_port"), &ChartHandler{})

	return handlers.MainHandler(updates, bot, db)
}
