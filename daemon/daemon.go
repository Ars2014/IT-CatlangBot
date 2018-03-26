package daemon

import (
	"errors"
	"log"
	"os"
	
	"github.com/Ars2014/IT-CatlangBot/dbHelper"
	"github.com/Ars2014/IT-CatlangBot/handlers"
	"github.com/Syfaro/telegram-bot-api"
)

func NewBot() (*tgbotapi.BotAPI, error) {
	token := os.Getenv("token")
	if token == "" {
		return nil, errors.New("no 'token' were specified in .env file")
	}
	bot, err := tgbotapi.NewBotAPI(os.Getenv("token"))
	if err != nil {
		return nil, err
	}
	if debug := os.Getenv("debug"); debug == "true" {
		bot.Debug = true
	}
	return bot, nil
}

func Run() error {
	bot, err := NewBot()
	if err != nil {
		return err
	}
	
	db, err := dbHelper.SetupDB()
	if err != nil {
		return err
	}
	
	log.Printf("[Daemon] Authorized on account %s", bot.Self.UserName)
	
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	
	updates, err := bot.GetUpdatesChan(u)
	
	if err != nil {
		return err
	}
	
	return handlers.MainHandler(updates, bot, db)
}
