package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"unicode/utf8"
)

type Cinema struct {
    ID               uint32 `json:"id"`
    Name             string `json:"name"`
    Year             uint16 `json:"year"`
    TypeNumber       int    `json:"typeNumber"`
    AgeRating        uint16 `json:"ageRating"`
    Description      string `json:"description"`
    ShortDescription string `json:"shortDescription"`
    MovieLength      uint16 `json:"movieLength,omitempty"`
    Countries        []struct {
        Name string `json:"name"`
    } `json:"countries"`
    Genres []struct {
        Name string `json:"name"`
    } `json:"genres"`
    BackDrop *struct {
        URL string `json:"url,omitempty"`
    } `json:"backdrop,omitempty"`
    Poster *struct {
        URL string `json:"url,omitempty"`
    } `json:"poster,omitempty"`
    Rating struct {
        Imdb float32 `json:"imdb"`
        Kp   float32 `json:"kp"`
    } `json:"rating"`
}


func Request(id string) (film string, picURL string, err error) {
	apiKey := os.Getenv("API_KEY")
	if apiKey == "" {
		return "", "", errors.New("переменная окружения API_KEY не задана")
	}
	//escapedName := url.QueryEscape(name)
	//apiURL := fmt.Sprintf("https://api.kinopoisk.dev/v1.4/movie/search?page=1&limit=1&query=%s", escapedName)
	//req, err := http.NewRequest("GET", apiURL, nil)
	url := "https://api.kinopoisk.dev/v1.4/movie/" + id
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", "", fmt.Errorf("ошибка создания запроса: %v", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-API-KEY", apiKey)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("ошибка при отправке запроса: %v", err)
	}
	defer res.Body.Close()

	// Проверка на статус код 403
	if res.StatusCode == http.StatusForbidden {
		return "", "", errors.New("закончился общий лимит на сегодня, возвращайтесь завтра")
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", "", fmt.Errorf("ошибка при чтении ответа: %v", err)
	}

	var cinema Cinema
	err = json.Unmarshal(body, &cinema)
	if err != nil {
		return "", "", fmt.Errorf("ошибка при разборе JSON: %v", err)
	}

	info := cinema	
	picURL = ""
	if cinema.Poster != nil && cinema.Poster.URL != "" {
		picURL = cinema.Poster.URL
	}else if cinema.BackDrop != nil && cinema.BackDrop.URL != "" {
		picURL = cinema.BackDrop.URL
	} else {
		picURL = "https://habrastorage.org/webt/bh/ex/-z/bhex-zst09dlgq-y2rjespcpp0c.png"
	}
	


	rating := cinema.Rating

	// собираем всё в одну переменную
	var sb strings.Builder

	if info.Name == ""{
		sb.WriteString("Фильм не найден, попробуйте снова")
	}else{
		// тип фильм аниме и т.д
	var TypeFilm string
	switch info.TypeNumber {
	case 1:
		TypeFilm = "Фильм"
	case 2:
		TypeFilm = "Сериал"
	case 3:
		TypeFilm = "Мультфильм"
	case 4:
		TypeFilm = "Аниме"
	case 5:
		TypeFilm = "Мультсериал"
	default:
		TypeFilm = ""
	}
	sb.WriteString(fmt.Sprintf("%s\n", TypeFilm))

	sb.WriteString(fmt.Sprintf("%v (%v) %v+\n", info.Name, info.Year, info.AgeRating))
	for _, genre := range info.Genres {
		sb.WriteString(fmt.Sprintf("#%v ", genre.Name))
	}
	sb.WriteString("\n")

	if info.MovieLength != 0 {
		hours := info.MovieLength / 60
		minutes := info.MovieLength % 60
		sb.WriteString(fmt.Sprintf("Длительность: %d ч %d мин\n", hours, minutes))
	}

	sb.WriteString("Страны: ")
	for i, country := range info.Countries {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(country.Name)
	}
	sb.WriteString("\n")

	// полное или краткое описание
	if utf8.RuneCountInString(info.Description) < 300 {
		sb.WriteString(fmt.Sprintf("Описание: %v\n", info.Description))
	} else if info.ShortDescription != "" {
		sb.WriteString(fmt.Sprintf("Краткое описание: %v\n", info.ShortDescription))
	} else {
		sb.WriteString("Описание отсутствует или слишком длинное")
	}
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("IMDb: %.1f КП: %.1f\n", rating.Imdb, rating.Kp))
	filmURL := fmt.Sprintf("<a href=\"https://www.kinopoisk.ru/film/%d/\">Смотреть подробнее</a>\n",
		info.ID)
	sb.WriteString(filmURL)
	}
	
	return sb.String(), picURL, nil
}
