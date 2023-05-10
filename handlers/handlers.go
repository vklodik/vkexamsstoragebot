package handlers

import (
	"database/sql"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
	"time"
	"vk-storage-bot/database"
)

type UserState struct {
	Step int
	Data []string
}

const (
	ServiceStep = iota
	LoginStep
	PasswordStep
)

func SetHandler(db *sql.DB, bot *tgbotapi.BotAPI, update tgbotapi.Update, updates tgbotapi.UpdatesChannel) {
	userID := update.Message.From.ID

	sent, err := bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Введите название сервиса:"))
	if err != nil {
		log.Println(err)
		return
	}

	states := make(map[int]*UserState)
	states[userID] = &UserState{
		Step: ServiceStep,
		Data: make([]string, 3),
	}

	for {
		select {
		case u := <-updates:
			if u.Message == nil || u.Message.Caption != "" || u.Message.Voice != nil || u.Message.Document != nil || u.Message.Photo != nil || u.Message.ForwardFrom != nil || len([]rune(u.Message.Text)) > 57 {
				continue
			}

			state := states[userID]

			switch state.Step {
			case ServiceStep:
				if u.Message.Text == "" || u.Message.From.ID != userID {
					continue
				}
				state.Data[0] = u.Message.Text
				if _, err := bot.Send(tgbotapi.NewEditMessageText(u.Message.Chat.ID, sent.MessageID, "Введите логин:")); err != nil {
					log.Println(err)
					return
				}
				go credentialsDelete(bot, u.Message, time.Nanosecond)
				state.Step = LoginStep
			case LoginStep:
				if u.Message.Text == "" || u.Message.From.ID != userID {
					continue
				}
				state.Data[1] = u.Message.Text
				if _, err := bot.Send(tgbotapi.NewEditMessageText(u.Message.Chat.ID, sent.MessageID, "Введите пароль:")); err != nil {
					log.Println(err)
					return
				}
				go credentialsDelete(bot, u.Message, time.Nanosecond)
				state.Step = PasswordStep
			case PasswordStep:
				if u.Message.Text == "" || u.Message.From.ID != userID {
					continue
				}
				state.Data[2] = u.Message.Text
				if err := database.Set(db, userID, state.Data[0], state.Data[1], state.Data[2]); err != nil {
					if _, err := bot.Send(tgbotapi.NewEditMessageText(u.Message.Chat.ID, sent.MessageID, "Ошибка при сохранении данных!")); err != nil {
						log.Println(err)
						return
					}
					return
				}

				if _, err := bot.Send(tgbotapi.NewEditMessageText(u.Message.Chat.ID, sent.MessageID, "Данные успешно добавлены")); err != nil {
					log.Println(err)
					return
				}
				go credentialsDelete(bot, u.Message, time.Nanosecond)

				delete(states, userID)
				return
			}
		}
	}
}

func GetHandler(db *sql.DB, bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	serviceNames, err := database.GetServices(db, message.From.ID)
	if err != nil {
		_, err := bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Ошибка при получении списка сервисов"))
		if err != nil {
			log.Println(err)
			return
		}
		log.Println(err)
		return
	}

	if len(serviceNames) == 0 {
		_, err := bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Список сервисов пуст"))
		if err != nil {
			log.Println(err)
			return
		}
		log.Println(err)
		return
	}

	var rows [][]tgbotapi.InlineKeyboardButton
	for _, service := range serviceNames {
		row := []tgbotapi.InlineKeyboardButton{tgbotapi.NewInlineKeyboardButtonData(service, "select_"+service)}
		rows = append(rows, row)
	}

	inlineKeyboardMarkup := tgbotapi.NewInlineKeyboardMarkup(rows...)
	msg := tgbotapi.NewMessage(message.Chat.ID, "Выберите сервис для получения реквизитов:")
	msg.ReplyMarkup = inlineKeyboardMarkup

	if _, err := bot.Send(msg); err != nil {
		log.Println(err)
		return
	}
}

func GetServiceHandler(db *sql.DB, bot *tgbotapi.BotAPI, serviceName string, query *tgbotapi.CallbackQuery) {
	login, pass, err := database.Get(db, query.From.ID, serviceName)
	if err != nil {
		if _, err := bot.AnswerCallbackQuery(tgbotapi.NewCallback(query.ID, "Ошибка при получении данных!")); err != nil {
			return
		}
		return
	}

	msg, err := bot.Send(tgbotapi.NewMessage(query.Message.Chat.ID, fmt.Sprintf("Логин: %s\nПароль: %s", login, pass)))
	if err != nil {
		return
	}
	go credentialsDelete(bot, &msg, time.Second*15)

	if _, err := bot.DeleteMessage(tgbotapi.NewDeleteMessage(query.Message.Chat.ID, query.Message.MessageID)); err != nil {
		return
	}
}

func DelHandler(db *sql.DB, bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	serviceNames, err := database.GetServices(db, message.From.ID)
	if err != nil {
		_, err := bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Ошибка при получении списка сервисов"))
		if err != nil {
			log.Println(err)
			return
		}
		log.Println(err)
		return
	}

	if len(serviceNames) == 0 {
		_, err := bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Список сервисов пуст"))
		if err != nil {
			log.Println(err)
			return
		}
		log.Println(err)
		return
	}

	var rows [][]tgbotapi.InlineKeyboardButton
	for _, service := range serviceNames {
		row := []tgbotapi.InlineKeyboardButton{tgbotapi.NewInlineKeyboardButtonData(service, "delete_"+service)}
		rows = append(rows, row)
	}

	inlineKeyboardMarkup := tgbotapi.NewInlineKeyboardMarkup(rows...)
	msg := tgbotapi.NewMessage(message.Chat.ID, "Выберите сервис для удаления:")
	msg.ReplyMarkup = inlineKeyboardMarkup

	if _, err := bot.Send(msg); err != nil {
		log.Println(err)
		return
	}
}

func DelServiceHandler(db *sql.DB, bot *tgbotapi.BotAPI, serviceName string, query *tgbotapi.CallbackQuery) {
	err := database.Del(db, query.From.ID, serviceName)
	if err != nil {
		if _, err := bot.AnswerCallbackQuery(tgbotapi.NewCallback(query.ID, "Ошибка при удалении данных!")); err != nil {
			return
		}
		return
	}

	if _, err := bot.Send(tgbotapi.NewMessage(query.Message.Chat.ID, "Данные успешно удалены!")); err != nil {
		return
	}

	if _, err := bot.DeleteMessage(tgbotapi.NewDeleteMessage(query.Message.Chat.ID, query.Message.MessageID)); err != nil {
		return
	}
}

func DefaultHandler(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	if _, err := bot.Send(
		tgbotapi.NewMessage(
			message.Chat.ID,
			"Введите одну из следующих команд:\n/set\n/get\n/del",
		),
	); err != nil {
		return
	}
}

func credentialsDelete(bot *tgbotapi.BotAPI, message *tgbotapi.Message, duration time.Duration) {
	time.AfterFunc(duration, func() {
		_, err := bot.DeleteMessage(tgbotapi.NewDeleteMessage(message.Chat.ID, message.MessageID))
		if err != nil {
			log.Println("Не удалось удалить сообщение!")
		}
	})
}
