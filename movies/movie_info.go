package movies

import (
	"fmt"
	"log"
	"strings"
	"unicode/utf8"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/luzhnov-aleksei/kinobot/api"
)

func HandleMovieSelection(bot *tgbotapi.BotAPI, update *tgbotapi.Update) {
	userID := update.CallbackQuery.From.ID
	selectedMovieID := update.CallbackQuery.Data

	// Получаем сохранённый список фильмов
	movies, exists := userMovieSelections[userID]
	if !exists {
		msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Произошла ошибка: список фильмов не найден.\nПопробуйте ввести новый запрос.")
		if _, err := bot.Send(msg); err != nil {
			log.Println("Ошибка при отправке сообщения:", err)
		}
		return
	}

	// Ищем выбранный фильм по ID
	var selectedMovie *api.Cinema
	for _, movie := range movies {
		if fmt.Sprint(movie.ID) == selectedMovieID {
			selectedMovie = &movie
			break
		}
	}

	if selectedMovie != nil {
		film, imageURL, err := FormatMovieInfo(selectedMovie)
		if err != nil {
			text := fmt.Sprintf("Произошла ошибка: %s", err)
			msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, text)
			if _, err := bot.Send(msg); err != nil {
				log.Println("Ошибка при отправке сообщения:", err)
			}
			return
		}

		// Отправка информации о фильме вместе с изображением
		photo := tgbotapi.NewInputMediaPhoto(tgbotapi.FileURL(imageURL))
		photo.Caption = film
		photo.ParseMode = "HTML"

		media := []interface{}{photo}
		mediaGroup := tgbotapi.MediaGroupConfig{
			ChatID: update.CallbackQuery.Message.Chat.ID,
			Media:  media,
		}

		_, err = bot.SendMediaGroup(mediaGroup)
		if err != nil {
			text := fmt.Sprintf("Произошла ошибка в отправке медиагруппы: %s", err)
			msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, text)
			if _, err := bot.Send(msg); err != nil {
				log.Println("Ошибка при отправке сообщения:", err)
			}

		}

	} else {
		msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Фильм не найден.")
		if _, err := bot.Send(msg); err != nil {
			log.Println("Ошибка при отправке сообщения:", err)
		}
	}
}

// тип: фильм, сериал, аниме и т.д.
func TypeFilm(TypeNumber int) string {
	switch TypeNumber {
	case 1:
		return "Фильм"
	case 2:
		return "Сериал"
	case 3:
		return "Мультфильм"
	case 4:
		return "Аниме"
	case 5:
		return "Мультсериал"
	default:
		return ""
	}
}

// Функция для форматирования информации о фильме
func FormatMovieInfo(movie *api.Cinema) (film string, picURL string, err error) {
	if movie.Poster != nil && movie.Poster.URL != "" {
		picURL = movie.Poster.URL
	} else if movie.BackDrop != nil && movie.BackDrop.URL != "" {
		picURL = movie.BackDrop.URL
	} else {
		picURL = "https://habrastorage.org/webt/bh/ex/-z/bhex-zst09dlgq-y2rjespcpp0c.png"
	}

	rating := movie.Rating

	// собираем всё в одну переменную
	var sb strings.Builder

	if movie.Name == "" {
		sb.WriteString("Фильм не найден, попробуйте снова")
	} else {
		typeFilm := TypeFilm(movie.TypeNumber)
		sb.WriteString(fmt.Sprintf("%s\n", typeFilm))

		sb.WriteString(fmt.Sprintf("%v (%v) %v+\n", movie.Name, movie.Year, movie.AgeRating))
		for _, genre := range movie.Genres {
			sb.WriteString(fmt.Sprintf("#%v ", genre.Name))
		}
		sb.WriteString("\n")

		if movie.MovieLength != 0 {
			hours := movie.MovieLength / 60
			minutes := movie.MovieLength % 60
			sb.WriteString(fmt.Sprintf("Длительность: %d ч %d мин\n", hours, minutes))
		}

		sb.WriteString("Страны: ")
		if len(movie.Countries) > 0 {
			for i, country := range movie.Countries {
				if i > 0 {
					sb.WriteString(", ")
				}
				sb.WriteString(country.Name)
			}
		} else {
			sb.WriteString("Страна не указана")
		}
		sb.WriteString("\n")

		// полное или краткое описание
		if utf8.RuneCountInString(movie.Description) < 350 {
			sb.WriteString(fmt.Sprintf("\nОписание: %v\n", movie.Description))
		} else if movie.ShortDescription != "" {
			sb.WriteString(fmt.Sprintf("\nКраткое описание: %v\n", movie.ShortDescription))
		} else {
			sb.WriteString("\nОписание слишком длинное.\nДля ознакомления с ним нажми <u>Смотреть подробнее</u>👇🏻\n")
		}
		sb.WriteString("\n")
		sb.WriteString(fmt.Sprintf("КП: %.1f IMDb: %.1f\n", rating.Kp, rating.Imdb))
		filmURL := fmt.Sprintf("<a href=\"https://www.kinopoisk.ru/film/%d/\">Смотреть подробнее</a>\n",
			movie.ID)
		sb.WriteString(filmURL)
	}

	return sb.String(), picURL, nil

}
