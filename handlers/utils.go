package handlers

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/Ars2014/IT-CatlangBot/models"
	"github.com/Syfaro/telegram-bot-api"
)

func getUsernameOrName(user *tgbotapi.User) string {
	if user.UserName == "" {
		name := strings.TrimSpace(fmt.Sprintf("%s %s", user.FirstName, user.LastName))

		if len(name) == 0 {
			return strconv.Itoa(user.ID)
		} else {
			return name
		}
	}

	return "@" + user.UserName
}

func getUsernameOrNameFromChat(chat *tgbotapi.Chat) string {
	if chat.UserName == "" {
		name := strings.TrimSpace(fmt.Sprintf("%s %s", chat.FirstName, chat.LastName))

		if len(name) == 0 {
			return strconv.Itoa(int(chat.ID))
		} else {
			return name
		}
	}

	return "@" + chat.UserName
}

func createLanguagesList(user *models.User) string {
	var languages []string

	userLanguages := user.Languages
	sort.Slice(userLanguages, func(i, j int) bool {
		return strings.ToLower(userLanguages[i].Name) < strings.ToLower(userLanguages[j].Name)
	})

	for _, lang := range userLanguages {
		languages = append(languages, lang.Name)
	}

	return strings.Join(languages, ", ")
}

func getNumEnding(number int, endings ...string) string {
	num := number % 100
	if 11 <= num && num <= 19 {
		return endings[2]
	}
	num %= 10
	if num == 1 {
		return endings[0]
	} else if 2 <= num && num <= 4 {
		return endings[1]
	}
	return endings[2]
}
