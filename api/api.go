package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"sort"
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

// Сортировка по рейтингу KП
type ByKpRating []Cinema

func (a ByKpRating) Len() int           { return len(a) }
func (a ByKpRating) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByKpRating) Less(i, j int) bool { return a[i].Rating.Kp > a[j].Rating.Kp }

// Запрос списка фильмов по названию
func RequestMovies(apiURL string, query string) ([]Cinema, error) {
	apiKey := os.Getenv("API_KEY")
	if apiKey == "" {
		return nil, errors.New("переменная окружения API_KEY не задана")
	}
	escapedName := url.QueryEscape(query)
	fullURL := fmt.Sprintf("%s?page=1&limit=8&query=%s", apiURL, escapedName)
	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания запроса: %v", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-API-KEY", apiKey)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ошибка при отправке запроса: %v", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("ошибка при чтении ответа: %v", err)
	}

	var results struct {
		Movies []Cinema `json:"docs"`
	}
	err = json.Unmarshal(body, &results)
	if err != nil {
		return nil, fmt.Errorf("ошибка при разборе JSON: %v", err)
	}

	// Сортировка фильмов по рейтингу KП по убыванию
	sort.Sort(ByKpRating(results.Movies))

	return results.Movies, nil
}
