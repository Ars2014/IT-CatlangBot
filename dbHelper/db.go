package dbHelper

import (
	"encoding/json"
	"errors"
	"log"
	"os"

	"github.com/Ars2014/IT-CatlangBot/models"
	"github.com/boltdb/bolt"
)

var (
	RootBucket      = []byte("DB")
	LanguagesBucket = []byte("LANGUAGES")
	UsersBucket     = []byte("USERS")
)

func SetupDB() (*bolt.DB, error) {
	dbFilename := os.Getenv("db")
	if dbFilename == "" {
		return nil, errors.New("no 'db' were specified in .env file")
	}

	db, err := bolt.Open(dbFilename, 0600, nil)
	if err != nil {
		return nil, err
	}

	err = db.Update(func(tx *bolt.Tx) error {
		root, err := tx.CreateBucketIfNotExists(RootBucket)
		if err != nil {
			return err
		}

		_, err = root.CreateBucketIfNotExists(LanguagesBucket)
		if err != nil {
			return err
		}

		_, err = root.CreateBucketIfNotExists(UsersBucket)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	log.Println("[DB] Setup done.")
	return db, nil
}

func AddLanguage(db *bolt.DB, language models.Language) error {
	var id uint64
	err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(RootBucket).Bucket(LanguagesBucket)

		id, _ = b.NextSequence()

		buf, err := json.Marshal(language)
		if err != nil {
			return err
		}

		return b.Put(ItoB(int(id)), buf)
	})

	if err == nil {
		log.Println("[DB] New language was successfully added.")
	}

	return err
}

func GetLanguages(db *bolt.DB) ([]models.Language, error) {
	var languages []models.Language

	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(RootBucket).Bucket(LanguagesBucket)

		return b.ForEach(func(key, value []byte) error {
			var language models.Language

			err := json.Unmarshal(value, &language)
			if err != nil {
				return err
			}

			languages = append(languages, language)
			return nil
		})
	})

	if err != nil {
		return nil, err
	}

	return languages, nil
}

func RemoveLanguage(db *bolt.DB, language models.Language) (bool, error) {
	var founded bool

	err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(RootBucket).Bucket(UsersBucket)
		return b.ForEach(func(key, value []byte) error {
			var user models.User
			err := json.Unmarshal(value, &user)
			if err != nil {
				return err
			}

			langIndex := -1
			for index, lang := range user.Languages {
				if lang.Name == language.Name {
					langIndex = index
					break
				}
			}

			if langIndex != -1 {
				user.Languages = append(user.Languages[:langIndex], user.Languages[langIndex+1:]...)
			}

			buf, err := json.Marshal(user)
			if err != nil {
				return err
			}

			return b.Put(ItoB(user.ID), buf)
		})
	})
	if err != nil {
		return founded, err
	}

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(RootBucket).Bucket(LanguagesBucket)
		var id []byte
		b.ForEach(func(key, value []byte) error {
			var lang models.Language
			err := json.Unmarshal(value, &lang)
			if err != nil {
				return err
			}
			if string(lang.Name) == language.Name {
				id = key
				founded = true
			}
			return nil
		})
		if founded {
			return b.Delete(id)
		}
		return nil
	})

	if err == nil && founded {
		log.Println("[DB] Language was successfully removed.")
	}

	return founded, err
}

func AddOrUpdateUser(db *bolt.DB, user models.User) error {
	err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(RootBucket).Bucket(UsersBucket)
		buf, err := json.Marshal(user)
		if err != nil {
			return err
		}
		return b.Put(ItoB(user.ID), buf)
	})

	if err == nil {
		log.Println("[DB] User was successfully added/updated.")
	}

	return err
}

func GetUser(db *bolt.DB, id int) (*models.User, error) {
	var user *models.User

	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(RootBucket).Bucket(UsersBucket)
		value := b.Get(ItoB(id))
		if value == nil {
			user = nil
			return nil
		}
		return json.Unmarshal(value, &user)
	})

	if err != nil {
		return nil, err
	}

	return user, nil
}

func CheckUserExists(db *bolt.DB, id int) (bool, error) {
	user, err := GetUser(db, id)
	if user != nil {
		return true, nil
	}
	return false, err
}

func AddUserIfNotExists(db *bolt.DB, id int) error {
	if check, err := CheckUserExists(db, id); err == nil {
		if !check {
			AddOrUpdateUser(db, models.User{id, []models.Language{}})
		}
		return nil
	} else {
		return err
	}
}

func GetUsers(db *bolt.DB) ([]models.User, error) {
	var users []models.User

	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(RootBucket).Bucket(UsersBucket)
		return b.ForEach(func(key, value []byte) error {
			var user models.User
			err := json.Unmarshal(value, &user)
			if err != nil {
				return err
			}
			users = append(users, user)
			return nil
		})
	})

	if err != nil {
		return nil, err
	}

	return users, nil
}

func RemoveUser(db *bolt.DB, id int) error {
	err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(RootBucket).Bucket(UsersBucket)
		return b.Delete(ItoB(id))
	})

	if err == nil {
		log.Println("[DB] User was successfully removed.")
	}

	return err
}
