package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/luzhnov-aleksei/kinobot/limiter"
	"github.com/luzhnov-aleksei/kinobot/movies"
)

func main() {
	botKey := os.Getenv("BOT_KEY")
	if botKey == "" {
		log.Fatal("BOT_KEY environment variable is not set")
	}

	bot, err := tgbotapi.NewBotAPI(botKey)
	if err != nil {
		log.Fatalf("Failed to authorize bot. Error: %v. This might be due to VPN issues.", err)
	}

	bot.Debug = false // Режим отладки
	log.Printf("Authorized on account %s", bot.Self.UserName)
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil {
			userID := update.Message.From.ID
			firstName := update.Message.From.FirstName
			username := update.Message.From.UserName

			if firstName == "" {
				firstName = "друг"
			}

			if username == "" {
				username = "username отсутствует"
			}

			if update.Message.NewChatMembers != nil {
				for _, member := range update.Message.NewChatMembers {
					log.Printf("New user authorized: %s (@%s)", member.FirstName, member.UserName)
				}
			}
			// Проверка на лимит сообщений
			if !limiter.CanSendMessage(userID) {
				log.Printf("Пользователь [%d] с username [@%s] превысил лимит сообщений", userID, username)
				msg := tgbotapi.NewMessage(update.Message.Chat.ID,
					"Вы превысили лимит сообщений на сегодня. Попробуйте снова завтра.")
				if _, err := bot.Send(msg); err != nil {
					log.Println("Ошибка при отправке сообщения из-за лимита на пользователя:", err)
				}
				continue
			}

			limiter.IncrementMessageCount(userID)

			// Общая часть сообщения
			commonMsg := "🤖 Это кинобот-помощник для создания списка фильмов и сериалов, которые ты планируешь посмотреть.\n\n" +
				"✏️ Просто напиши боту запрос, выбери нужный фильм и бот выдаст информацию о нем.\n\n" +
				"🔎 Поиск работает по названию, жанру, году. Также можно это комбинировать\n\n" +
				"📽️ Бот может искать всё, что есть на Кинопоиске: фильмы, мультфильмы, сериалы, аниме и т.д.\n\n" +
				"📝 Личку бота можно использовать как записную книгу с фильмами.\n\n" +
				"💬 Или добавь бота в любой чат, дай ему админку, и он будет присылать туда фильмы по вашим запросам👍\n\n" +
				"🤔 Если возникнут вопросы или проблемы с ботом, то напиши разработчику @luzhnov_aleksei"

			var msgText string

			// Обработка команд /start и /help
			if update.Message.Text == "/start" {
				log.Printf("/start от пользователя [%d] с username [@%s]: %s", userID, username, update.Message.Text)
				msgText = fmt.Sprintf("Привет, %s👋👋👋\n\n", firstName) + commonMsg
			} else if update.Message.Text == "/help" {
				msgText = commonMsg
			} else {
				log.Printf("Получено сообщение от пользователя [%d] с username [@%s]: %s", userID, username, update.Message.Text)
				// Отправляем анимацию загрузки
				animation := tgbotapi.NewAnimation(update.Message.Chat.ID, tgbotapi.FileURL("https://media1.tenor.com/m/RVvnVPK-6dcAAAAd/reload-cat.gif"))
				animationMsg, err := bot.Send(animation)
				if err != nil {
					log.Printf("Не удалось отправить GIF: %v", err)
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "🔄 Идет поиск... Пожалуйста, подождите.")
					if _, err := bot.Send(msg); err != nil {
						log.Println("Ошибка при отправке сообщения из-за отсутствия gif:", err)
					}
					movies.HandleMovieSearch(bot, &update)
				} else {
					// Обрабатываем запрос фильма после успешной отправки GIF
					movies.HandleMovieSearch(bot, &update)

					// Удаляем GIF после обработки запроса
					deleteMessage := tgbotapi.NewDeleteMessage(update.Message.Chat.ID, animationMsg.MessageID)
					_, deleteErr := bot.Request(deleteMessage)
					if deleteErr != nil {
						log.Println("Ошибка при удалении GIF сообщения:", deleteErr)
					}
				}
			}

			// Отправка сообщения для /start или /help
			if msgText != "" {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, strings.TrimSpace(msgText))
				if _, err := bot.Send(msg); err != nil {
					log.Println("Ошибка при отправке /start /help сообщения:", err)
				}
			}

		} else if update.CallbackQuery != nil {
			// Обработка выбора фильма из списка
			movies.HandleMovieSelection(bot, &update)
		}
	}
}
