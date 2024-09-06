package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/luzhnov-aleksei/kinobot_otus/api"
	"github.com/luzhnov-aleksei/kinobot_otus/limiter"
)

func BotAuthorization(botKey string) (*tgbotapi.BotAPI, error) {
	if botKey == "" {
		return nil, fmt.Errorf("переменная окружения BotKey не задана")
	}

	bot, err := tgbotapi.NewBotAPI(botKey)
	if err != nil {
		return nil, err
	}

	return bot, nil
}

func main() {
	botKey := os.Getenv("BOT_KEY")

	bot, err := BotAuthorization(botKey)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = false // потом врубить

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil {
			userID := update.Message.From.ID
			userName := update.Message.From.FirstName

			if userName == "" {
				userName = "друг"
			}

			// Проверка на лимит сообщений
			if !limiter.CanSendMessage(userID) {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID,
					"Вы превысили лимит сообщений на сегодня. Попробуйте снова завтра.")
				bot.Send(msg)
				continue
			}

			limiter.IncrementMessageCount(userID)

			// Общая часть сообщения
			commonMsg := "🤖 Это кинобот-помощник для создания списка фильмов и сериалов, которые ты планируешь посмотреть.\n\n" +
				"✏️ Просто напиши боту название фильма, и он выдаст информацию о нем.\n\n" +
				"📝 Личку бота можно использовать как записную книгу с фильмами.\n\n" +
				"💬 Или добавь бота в любой чат, дай ему админку, и он будет присылать туда фильмы по вашим запросам👍\n\n" +
				"🤔 Если возникнут вопросы или проблемы с ботом, то напиши разработчику @luzhnov_aleksei"

			var msgText string

			// Обработка команд /start и /help
			if update.Message.Text == "/start" {
				msgText = fmt.Sprintf("Привет, %s👋👋👋\n\n", userName) + commonMsg
			} else if update.Message.Text == "/help" {
				msgText = commonMsg
			} else {
				// Если не /start и не /help, запрос к API
				film, imageURL, err := api.Request(update.Message.Text)
				if err != nil {
					text := fmt.Sprintf("Произошла ошибка: %s", err)
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, text)
					bot.Send(msg)
					continue
				}

				// Отправка информации о фильме
				photo := tgbotapi.NewInputMediaPhoto(tgbotapi.FileURL(imageURL))
				photo.Caption = film
				photo.ParseMode = "HTML"

				media := []interface{}{photo}
				mediaGroup := tgbotapi.MediaGroupConfig{
					ChatID: update.Message.Chat.ID,
					Media:  media,
				}

				_, err = bot.SendMediaGroup(mediaGroup)
				if err != nil {
					text := fmt.Sprintf("Произошла ошибка в сборке сообщения: %s", err)
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, text)
					bot.Send(msg)
				}

				continue
			}

			// Отправка сообщения
			if msgText != "" {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, strings.TrimSpace(msgText))
				bot.Send(msg)
			}
		}
	}
}