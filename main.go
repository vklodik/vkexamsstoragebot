package main

import (
	"database/sql"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/ilyakaznacheev/cleanenv"
	_ "github.com/lib/pq"
	"log"
	"strings"
	"vk-storage-bot/config"
	"vk-storage-bot/handlers"
)

func main() {
	var cfg config.Config

	err := cleanenv.ReadConfig("config.yml", &cfg)
	if err != nil {
		log.Fatal(err)
	}

	connStr := fmt.Sprintf(
		"user=%s password=%s dbname=%s sslmode=disable",
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Name,
	)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

	bot, err := tgbotapi.NewBotAPI(cfg.Telegram.Token)
	if err != nil {
		log.Fatal(err)
	}

	bot.Debug = false

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		log.Fatal(err)
	}

	for update := range updates {
		if update.CallbackQuery != nil {
			parts := strings.Split(update.CallbackQuery.Data, "_")

			switch parts[0] {
			case "select":
				handlers.GetServiceHandler(db, bot, parts[1], update.CallbackQuery)
			case "delete":
				handlers.DelServiceHandler(db, bot, parts[1], update.CallbackQuery)
			}

		}

		if update.Message == nil {
			continue
		}

		if !update.Message.IsCommand() {
			continue
		}

		switch update.Message.Command() {
		case "set":
			handlers.SetHandler(db, bot, update, updates)
		case "get":
			handlers.GetHandler(db, bot, update.Message)
		case "del":
			handlers.DelHandler(db, bot, update.Message)
		default:
			handlers.DefaultHandler(bot, update.Message)
		}

	}
}
