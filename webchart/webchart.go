package webchart

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"path"
	"runtime"

	"github.com/Ars2014/IT-CatlangBot/dbHelper"
	"github.com/Ars2014/IT-CatlangBot/models"
	"github.com/boltdb/bolt"
)

func ChartPage(db *bolt.DB) (string, error) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return "", errors.New("could not find module file location")
	}
	filename := path.Join(path.Dir(file), "webchart.html")

	pageTemplate, err := template.ParseFiles(filename)
	if err != nil {
		return "", err
	}

	var users []models.User
	stat := make(map[string]int)

	usersInfo, err := dbHelper.GetUsers(db)
	if err != nil {
		return "", err
	}

	for _, user := range usersInfo {
		users = append(users, user)
		for _, lang := range user.Languages {
			stat[lang.Name] = stat[lang.Name] + 1
		}
	}

	data := [][]interface{}{{"Язык программирования", "Количество людей", map[string]string{"role": "style"}, map[string]string{"role": "annotation"}}}

	usersCount := len(users)
	for key, value := range stat {
		percent := (value / usersCount) * 100
		data = append(data, []interface{}{key, value, ColorFromString(key), fmt.Sprintf("%d%%", percent)})
	}

	marshaledData, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := pageTemplate.ExecuteTemplate(&buf, "Chart", template.JS(string(marshaledData))); err != nil {
		return "", err
	}

	return buf.String(), nil
}
