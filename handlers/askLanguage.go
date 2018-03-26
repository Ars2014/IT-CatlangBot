package handlers

import (
	"sort"
	"strings"
	
	"github.com/Ars2014/IT-CatlangBot/dbHelper"
	"github.com/Syfaro/telegram-bot-api"
	"github.com/boltdb/bolt"
)

func askLanguage(message *tgbotapi.Message, bot *tgbotapi.BotAPI, db *bolt.DB, handleException bool) error {
	var buttons [][]tgbotapi.InlineKeyboardButton
	
	languages, err := dbHelper.GetLanguages(db)
	if err != nil {
		return err
	}
	
	if len(languages) <= 0 {
		msg := tgbotapi.NewMessage(message.Chat.ID, "Ни один язык программирования был добавлен. Обратитесь к @Ars2013 за помощью.")
		msg.ReplyToMessageID = message.MessageID
		_, err = bot.Send(msg)
		return err
	}
	
	sort.Slice(languages, func(i, j int) bool {
		return strings.ToLower(languages[i].Name) < strings.ToLower(languages[j].Name)
	})
	
	for i := 0; i < len(languages); i += InlineKeyboardSize {
		var row []tgbotapi.InlineKeyboardButton
		
		end := i + InlineKeyboardSize
		
		if end > len(languages) {
			end = len(languages)
		}
		
		for _, lang := range languages[i:end] {
			row = append(row, tgbotapi.NewInlineKeyboardButtonData(lang.Name, lang.Name))
		}
		
		buttons = append(buttons, row)
	}
	markup := tgbotapi.NewInlineKeyboardMarkup(buttons...)
	
	msg := tgbotapi.NewMessage(int64(message.From.ID), "На каком языке прогроаммирования пишешь?")
	msg.ReplyMarkup = markup
	
	_, err = bot.Send(msg)
	
	if err != nil {
		goto ErrorCheck
	}
	
	if message.Chat.Type != "private" {
		msg = tgbotapi.NewMessage(message.Chat.ID, "Я отправил тебе вопрос личным сообщением.")
		msg.ReplyToMessageID = message.MessageID
		_, err = bot.Send(msg)
	}

ErrorCheck:
	if err != nil {
		if handleException {
			msg = tgbotapi.NewMessage(message.Chat.ID, "Произошла ошибка! Я не могу тебе написать в личку.\nНапиши мне в ЛС и я смогу тебе отвечать.")
			msg.ReplyToMessageID = message.MessageID
			_, err = bot.Send(msg)
		}
		return err
	}
	
	return nil
}
