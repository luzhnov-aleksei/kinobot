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

	bot.Debug = false
	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	// Обработка всех обновлений
	for update := range updates {
		handleUpdate(bot, update)
	}
}

// Разделение обработки обновлений
func handleUpdate(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	if update.Message != nil {
		handleMessage(bot, update)
	} else if update.CallbackQuery != nil {
		// Обработка выбора фильма из списка
		movies.HandleMovieSelection(bot, &update)
	}
}

// Разделение обработки сообщений
func handleMessage(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	userID := update.Message.From.ID
	firstName := update.Message.From.FirstName
	username := update.Message.From.UserName

	if firstName == "" {
		firstName = "друг"
	}

	if username == "" {
		username = "username отсутствует"
	}

	// Обработка новых участников чата
	if update.Message.NewChatMembers != nil {
		for _, member := range update.Message.NewChatMembers {
			log.Printf("New user authorized: %s (@%s)", member.FirstName, member.UserName)
		}
	}

	// Проверка на лимит сообщений
	if !limiter.CanSendMessage(userID) {
		handleMessageLimit(bot, update.Message.Chat.ID, int(userID), username)
		return
	}

	limiter.IncrementMessageCount(userID)

	// Обработка команд /start и /help
	switch update.Message.Text {
	case "/start":
		handleStartCommand(bot, update, firstName)
	case "/help":
		handleHelpCommand(bot, update)
	default:
		handleMovieSearch(bot, update)
	}
}

// Обработка превышения лимита сообщений
func handleMessageLimit(bot *tgbotapi.BotAPI, chatID int64, userID int, username string) {
	log.Printf("Пользователь [%d] с username [@%s] превысил лимит сообщений", userID, username)
	msg := tgbotapi.NewMessage(chatID, "Вы превысили лимит сообщений на сегодня. Попробуйте снова завтра.")
	if _, err := bot.Send(msg); err != nil {
		log.Println("Ошибка при отправке сообщения из-за лимита на пользователя:", err)
	}
}

// Обработка команды /start
func handleStartCommand(bot *tgbotapi.BotAPI, update tgbotapi.Update, firstName string) {
	commonMsg := getCommonMessage()
	msgText := fmt.Sprintf("Привет, %s👋👋👋\n\n", firstName) + commonMsg
	sendMessage(bot, update.Message.Chat.ID, msgText)
}

// Обработка команды /help
func handleHelpCommand(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	commonMsg := getCommonMessage()
	sendMessage(bot, update.Message.Chat.ID, commonMsg)
}

// Получение общего сообщения
func getCommonMessage() string {
	return "🤖 Это кинобот-помощник для создания списка фильмов и сериалов, которые ты планируешь посмотреть.\n\n" +
		"✏️ Просто напиши боту запрос, выбери нужный фильм и бот выдаст информацию о нем.\n\n" +
		"🔎 Поиск работает по названию, жанру, году. Также можно это комбинировать\n\n" +
		"📽️ Бот может искать всё, что есть на Кинопоиске: фильмы, мультфильмы, сериалы, аниме и т.д.\n\n" +
		"📝 Личку бота можно использовать как записную книгу с фильмами.\n\n" +
		"💬 Или добавь бота в любой чат, дай ему админку, и он будет присылать туда фильмы по вашим запросам👍\n\n" +
		"🤔 Если возникнут вопросы или проблемы с ботом, то напиши разработчику @luzhnov_aleksei"
}

// Обработка поиска фильмов
func handleMovieSearch(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	log.Printf("Получено сообщение от пользователя [%d] с username [@%s]: %s", update.Message.From.ID, update.Message.From.UserName, update.Message.Text)
	animation := tgbotapi.NewAnimation(update.Message.Chat.ID, tgbotapi.FileURL("https://media1.tenor.com/m/RVvnVPK-6dcAAAAd/reload-cat.gif"))
	animationMsg, err := bot.Send(animation)
	if err != nil {
		log.Printf("Не удалось отправить GIF: %v", err)
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "🔄 Идет поиск... Пожалуйста, подождите.")
		if _, err := bot.Send(msg); err != nil {
			log.Println("Ошибка при отправке сообщения из-за отсутствия gif:", err)
		}
		movies.HandleMovieSearch(bot, &update)
		return
	}

	// Обрабатываем запрос фильма после успешной отправки GIF
	movies.HandleMovieSearch(bot, &update)

	// Удаляем GIF после обработки запроса
	deleteMessage := tgbotapi.NewDeleteMessage(update.Message.Chat.ID, animationMsg.MessageID)
	if _, deleteErr := bot.Request(deleteMessage); deleteErr != nil {
		log.Println("Ошибка при удалении GIF сообщения:", deleteErr)
	}
}

// Функция для отправки сообщений
func sendMessage(bot *tgbotapi.BotAPI, chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, strings.TrimSpace(text))
	if _, err := bot.Send(msg); err != nil {
		log.Println("Ошибка при отправке сообщения:", err)
	}
}
