package movies

import (
	"fmt"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/luzhnov-aleksei/kinobot/api"
)

var UserMovieSelections = make(map[int64][]api.Cinema)
var userPreviousMessages = make(map[int64]int)
var userPreviousLists = make(map[int64]int)

// URL API можно передать как параметр или установить глобально
const apiURL = "https://api.kinopoisk.dev/v1.4/movie/search"

// Обработчик поиска фильмов
func HandleMovieSearch(bot *tgbotapi.BotAPI, update *tgbotapi.Update) {
	userID := update.Message.From.ID

	// Удаляем предыдущее сообщение пользователя, если оно существует
	if msgID, ok := userPreviousMessages[userID]; ok {
		deleteMsg := tgbotapi.NewDeleteMessage(update.Message.Chat.ID, msgID)
		if _, err := bot.Request(deleteMsg); err != nil {
			log.Println("Ошибка при удалении сообщения:", err)
		}
	}

	// Удаляем предыдущее сообщение со списком фильмов, если оно существует
	if listID, ok := userPreviousLists[userID]; ok {
		deleteMsg := tgbotapi.NewDeleteMessage(update.Message.Chat.ID, listID)
		if _, err := bot.Request(deleteMsg); err != nil {
			log.Println("Ошибка при удалении сообщения:", err)
		}
	}

	// Получаем список фильмов по запросу
	movies, err := api.RequestMovies(apiURL, update.Message.Text)
	if err != nil {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Произошла ошибка: %s", err))
		if _, err := bot.Send(msg); err != nil {
			log.Println("Ошибка при отправке сообщения:", err)
		}
		return
	}

	if len(movies) == 0 {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Фильм не найден, попробуйте другой запрос")
		if _, err := bot.Send(msg); err != nil {
			log.Println("Ошибка при отправке сообщения:", err)
		}
		return
	}

	// Сохраняем список фильмов
	UserMovieSelections[userID] = movies
	var countryName string

	// Формируем inline-кнопки
	var buttons [][]tgbotapi.InlineKeyboardButton
	for _, movie := range movies {
		if len(movie.Countries) > 0 {
			countryName = movie.Countries[0].Name
		} else {
			countryName = "Страна не указана"
		}
		button := tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("%s (%s, %s, %d)", movie.Name, TypeFilm(movie.TypeNumber), countryName, movie.Year), fmt.Sprint(movie.ID))
		buttons = append(buttons, tgbotapi.NewInlineKeyboardRow(button))
	}

	// Отправляем сообщение с выбором фильмов
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Выберите фильм:")
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(buttons...)
	sentMsg, _ := bot.Send(msg)

	// Сохраняем ID отправленного сообщения и ID запроса пользователя
	userPreviousMessages[userID] = update.Message.MessageID
	userPreviousLists[userID] = sentMsg.MessageID
}
