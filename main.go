package main

import (
	"fmt"
	"log"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/luzhnov-aleksei/kinobot/api"
	"github.com/luzhnov-aleksei/kinobot/config"
)

func main() {
	bot, err := tgbotapi.NewBotAPI(config.BotKey)
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
			if update.Message.Text == "/start" {
				userName := update.Message.From.FirstName
				if userName == "" {
					userName = "друг"
				}
				var sb strings.Builder
				sb.WriteString(fmt.Sprintf("Привет, %s.\n", userName))
				sb.WriteString("Это бот помощник для создания списка твоих любимых фильмов\n")
				sb.WriteString("или фильмов которые ты хочешь посмотреть.\n\n")
				sb.WriteString("Просто напиши боту название фильма и он выдаст инфорацию о нем.")
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, sb.String())
				bot.Send(msg)
			} else {
				//log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

				msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
				msg.ReplyToMessageID = update.Message.MessageID
				film, imageURL, err := api.Request(update.Message.Text)
				if err != nil {
					text := fmt.Sprintf("Произошла ошибка: %s", err)
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, text)
					bot.Send(msg)
				} else {
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, film)
					bot.Send(msg)
					if imageURL != "" {
						file := tgbotapi.FileURL(imageURL)
						photo := tgbotapi.NewPhoto(update.Message.Chat.ID, file)
						bot.Send(photo)
					}

				}
			}
		}
	}

}
