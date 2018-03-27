package handlers

import (
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/Ars2014/IT-CatlangBot/dbHelper"
	"github.com/Ars2014/IT-CatlangBot/models"
	"github.com/Syfaro/telegram-bot-api"
	"github.com/boltdb/bolt"
)

const InlineKeyboardSize = 3

func MainHandler(updateChannel tgbotapi.UpdatesChannel, bot *tgbotapi.BotAPI, db *bolt.DB) error {
	var (
		errChannel         = make(chan error)
		startChannel       = make(chan *tgbotapi.Message)
		langChannel        = make(chan *tgbotapi.Message)
		clearChannel        = make(chan *tgbotapi.Message)
		langForwardChannel = make(chan *tgbotapi.Message)
		catlangChannel     = make(chan *tgbotapi.Message)
		addLangChannel     = make(chan *tgbotapi.Message)
		rmLangChannel      = make(chan *tgbotapi.Message)
		statChannel        = make(chan *tgbotapi.Message)
		allStatChannel     = make(chan *tgbotapi.Message)
		clbChannel         = make(chan *tgbotapi.CallbackQuery)
	)

	go startHandler(startChannel, bot, db, errChannel)
	go langHandler(langChannel, bot, db, errChannel)
	go clearHandler(clearChannel, bot, db, errChannel)
	go langForwardHandler(langForwardChannel, bot, db, errChannel)
	go catlangHandler(catlangChannel, bot, db, errChannel)
	go addLanguageHandler(addLangChannel, bot, db, errChannel)
	go removeLanguageHandler(rmLangChannel, bot, db, errChannel)
	go statHandler(statChannel, bot, db, errChannel)
	go allStatHandler(allStatChannel, bot, db, errChannel)
	go inlineKeyboardHandler(clbChannel, bot, db, errChannel)

	for {
		select {
		case update := <-updateChannel:
			if update.CallbackQuery != nil {
				clbChannel <- update.CallbackQuery
				continue
			}
			if update.Message == nil {
				continue
			}

			msg := update.Message

			switch msg.Command() {
			case "start":
				startChannel <- msg
			case "lang":
				langChannel <- msg
			case "clear":
				clearChannel <- msg
			case "catlang":
				catlangChannel <- msg
			case "stat":
				statChannel <- msg
			case "allstat":
				allStatChannel <- msg
			case "addlang":
				addLangChannel <- msg
			case "rmlang":
				rmLangChannel <- msg
			}

			if msg.ForwardFrom != nil && msg.Chat.Type == "private" {
				langForwardChannel <- msg
			}

		case err := <-errChannel:
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func startHandler(msgChan chan *tgbotapi.Message, bot *tgbotapi.BotAPI, db *bolt.DB, errChan chan error) {
	for {
		msg := <-msgChan
		if msg.Chat.Type == "private" {
			errChan <- askLanguage(msg, bot, db, false)
		}
	}
}

func langHandler(msgChan chan *tgbotapi.Message, bot *tgbotapi.BotAPI, db *bolt.DB, errChan chan error) {
	for {
		message := <-msgChan

		var user *tgbotapi.User
		if message.ReplyToMessage != nil {
			user = message.ReplyToMessage.From
		} else {
			user = message.From
		}

		userInfo, err := dbHelper.GetUser(db, user.ID)
		if err != nil {
			errChan <- err
			continue
		}

		var text string
		if userInfo != nil && len(userInfo.Languages) > 0 {
			text = fmt.Sprintf("Пользователь %s пишет на: %s", getUsernameOrName(user), createLanguagesList(userInfo))
		} else {
			text = fmt.Sprintf("Пользователь %s не указал на каких ЯП пишет.", getUsernameOrName(user))
		}

		msg := tgbotapi.NewMessage(message.Chat.ID, text)
		msg.ReplyToMessageID = message.MessageID
		bot.Send(msg)
	}
}

func clearHandler(msgChan chan *tgbotapi.Message, bot *tgbotapi.BotAPI, db *bolt.DB, errChan chan error) {
	for {
		message := <-msgChan
		err := dbHelper.RemoveUser(db, message.From.ID)
		if err != nil {
			errChan <- err
			continue
		}
		msg := tgbotapi.NewMessage(int64(message.From.ID), "Информация о вас была очищена.")
		_, err = bot.Send(msg)
		errChan <- err
	}
}

func langForwardHandler(msgChan chan *tgbotapi.Message, bot *tgbotapi.BotAPI, db *bolt.DB, errChan chan error) {
	for {
		message := <-msgChan
		user := message.ForwardFrom

		userInfo, err := dbHelper.GetUser(db, user.ID)
		if err != nil {
			errChan <- err
			continue
		}

		var text string
		if userInfo != nil && len(userInfo.Languages) > 0 {
			text = fmt.Sprintf("Пользователь %s пишет на: %s", getUsernameOrName(user), createLanguagesList(userInfo))
		} else {
			text = fmt.Sprintf("Пользователь %s не указал на каких ЯП пишет.", getUsernameOrName(user))
		}

		msg := tgbotapi.NewMessage(message.Chat.ID, text)
		msg.ReplyToMessageID = message.MessageID
		bot.Send(msg)
	}
}

func catlangHandler(msgChan chan *tgbotapi.Message, bot *tgbotapi.BotAPI, db *bolt.DB, errChan chan error) {
	for {
		message := <-msgChan
		errChan <- askLanguage(message, bot, db, false)
	}
}

func addLanguageHandler(msgChan chan *tgbotapi.Message, bot *tgbotapi.BotAPI, db *bolt.DB, errChan chan error) {
	for {
		message := <-msgChan

		if strconv.Itoa(message.From.ID) != os.Getenv("admin") {
			msg := tgbotapi.NewMessage(message.Chat.ID, "Только админ может запускать данную команду.")
			msg.ReplyToMessageID = message.MessageID
			bot.Send(msg)
			continue
		}

		args := strings.TrimSpace(message.CommandArguments())

		if args == "" {
			msg := tgbotapi.NewMessage(message.Chat.ID, "Укажи название языка программирования как аргумент команды.")
			msg.ReplyToMessageID = message.MessageID
			bot.Send(msg)
			continue
		}

		languages, err := dbHelper.GetLanguages(db)
		if err != nil {
			errChan <- err
			continue
		}

		pos := -1
		for index, language := range languages {
			if language.Name == args {
				pos = index
			}
		}

		if pos == -1 {
			log.Printf("[Handler] Adding new language '%s'.", args)
			err = dbHelper.AddLanguage(db, models.Language{Name: args})
			if err != nil {
				errChan <- err
				continue
			}

			msg := tgbotapi.NewMessage(message.Chat.ID, "Новый язык программирования был добавлен.")
			msg.ReplyToMessageID = message.MessageID
			bot.Send(msg)
		} else {
			msg := tgbotapi.NewMessage(message.Chat.ID, "Данный язык программирования уже существует.")
			msg.ReplyToMessageID = message.MessageID
			bot.Send(msg)
		}
	}
}

func removeLanguageHandler(msgChan chan *tgbotapi.Message, bot *tgbotapi.BotAPI, db *bolt.DB, errChan chan error) {
	for {
		message := <-msgChan
		args := strings.TrimSpace(message.CommandArguments())

		if strconv.Itoa(message.From.ID) != os.Getenv("admin") {
			msg := tgbotapi.NewMessage(message.Chat.ID, "Только админ может запускать данную команду.")
			msg.ReplyToMessageID = message.MessageID
			bot.Send(msg)
			continue
		}

		founded, err := dbHelper.RemoveLanguage(db, models.Language{Name: args})
		if err != nil {
			errChan <- err
		}

		if !founded {
			msg := tgbotapi.NewMessage(message.Chat.ID, "Данный язык программирования не был найден.")
			msg.ReplyToMessageID = message.MessageID
			bot.Send(msg)
			continue
		}

		msg := tgbotapi.NewMessage(message.Chat.ID, "Язык программирования был удалён.")
		msg.ReplyToMessageID = message.MessageID
		bot.Send(msg)
	}

}

func statHandler(msgChan chan *tgbotapi.Message, bot *tgbotapi.BotAPI, db *bolt.DB, errChan chan error) {
	for {
		message := <-msgChan

		var users []models.User
		stat := make(map[string]int)

		usersInfo, err := dbHelper.GetUsers(db)
		if err != nil {
			errChan <- err
			continue
		}

		if len(usersInfo) == 0 {
			msg := tgbotapi.NewMessage(message.Chat.ID, "Ещё никто не ответил на вопрос.")
			msg.ReplyToMessageID = message.MessageID
			bot.Send(msg)
			continue
		}

		for _, user := range usersInfo {
			users = append(users, user)
			for _, lang := range user.Languages {
				stat[lang.Name] = stat[lang.Name] + 1
			}
		}

		var keys []string
		for k := range stat {
			keys = append(keys, k)
		}

		sort.Strings(keys)

		var maxKey string
		var maxValue int
		for _, key := range keys {
			value := stat[key]
			if value > maxValue {
				maxValue = value
				maxKey = key
			}
		}

		text := "Статистика:\n" +
			fmt.Sprintf("Всего %d %s на вопрос. ", len(users), getNumEnding(len(users),
				"пользователь ответил", "пользователя ответило", "пользователей ответили")) +
			fmt.Sprintf("Среди них самым популярным языком программирования является %s (%d %s).\n",
				maxKey, maxValue, getNumEnding(maxValue, "голос", "голоса", "голосов")) +
			"Статистика: " + os.Getenv("base_url") + os.Getenv("chart_url")

		msg := tgbotapi.NewMessage(message.Chat.ID, text)
		msg.ReplyToMessageID = message.MessageID
		bot.Send(msg)
	}
}

func allStatHandler(msgChan chan *tgbotapi.Message, bot *tgbotapi.BotAPI, db *bolt.DB, errChan chan error) {
	for {
		message := <-msgChan
		if message.Chat.Type != "private" {
			continue
		}
		msg := tgbotapi.NewMessage(message.Chat.ID, "Сбор информации о пользователях...\n"+
			"Это займёт некоторое время...")
		statMsg, _ := bot.Send(msg)

		users, err := dbHelper.GetUsers(db)
		if err != nil {
			errChan <- err
			continue
		}

		text := "Статистика всех пользователей:\n"
		for index, user := range users {
			index += 1
			languages := createLanguagesList(&user)

			userInfo, err := bot.GetChat(message.Chat.ChatConfig())
			if err != nil {
				text += fmt.Sprintf("%d) UID:%d - %s.", index, user.ID, languages)
			} else {
				text += fmt.Sprintf("%d) %s - %s.", index, getUsernameOrNameFromChat(&userInfo), languages)
			}
		}

		editMsg := tgbotapi.NewEditMessageText(message.Chat.ID, statMsg.MessageID, text)
		bot.Send(editMsg)
	}
}

func inlineKeyboardHandler(clbChan chan *tgbotapi.CallbackQuery, bot *tgbotapi.BotAPI, db *bolt.DB, errChan chan error) {
	for {
		callbackQuery := <-clbChan

		err := dbHelper.AddUserIfNotExists(db, callbackQuery.From.ID)
		if err != nil {
			errChan <- err
			continue
		}

		userInfo, err := dbHelper.GetUser(db, callbackQuery.From.ID)
		if err != nil {
			errChan <- err
			continue
		}

		removePos := -1
		for index, lang := range userInfo.Languages {
			if callbackQuery.Data == lang.Name {
				removePos = index
			}
		}

		if removePos == -1 {
			userInfo.Languages = append(userInfo.Languages, models.Language{Name: callbackQuery.Data})
			err = dbHelper.AddOrUpdateUser(db, *userInfo)
			if err != nil {
				errChan <- err
				continue
			}

			callbackQueryAnswer := tgbotapi.NewCallback(callbackQuery.ID, fmt.Sprintf("Выбрано: %s", callbackQuery.Data))
			bot.AnswerCallbackQuery(callbackQueryAnswer)
		} else {
			userInfo.Languages = append(userInfo.Languages[:removePos], userInfo.Languages[removePos+1:]...)
			if len(userInfo.Languages) > 0 {
				err = dbHelper.AddOrUpdateUser(db, *userInfo)
			} else {
				err = dbHelper.RemoveUser(db, userInfo.ID)
			}
			if err != nil {
				errChan <- err
				continue
			}
			callbackQueryAnswer := tgbotapi.NewCallback(callbackQuery.ID, fmt.Sprintf("Снят выбор с: %s", callbackQuery.Data))
			bot.AnswerCallbackQuery(callbackQueryAnswer)
		}
	}
}
