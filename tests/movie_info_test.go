package api

import (
	"testing"

	"github.com/luzhnov-aleksei/kinobot/api"
	"github.com/luzhnov-aleksei/kinobot/movies"
	"github.com/stretchr/testify/assert"
)


func TestTypeFilm(t *testing.T) {
	assert.Equal(t, "Фильм", movies.TypeFilm(1))
	assert.Equal(t, "Сериал", movies.TypeFilm(2))
	assert.Equal(t, "Мультфильм", movies.TypeFilm(3))
	assert.Equal(t, "Аниме", movies.TypeFilm(4))
	assert.Equal(t, "Мультсериал", movies.TypeFilm(5))
	assert.Equal(t, "", movies.TypeFilm(0))
}


func TestFormatMovieInfo(t *testing.T) {
	testMovie := &api.Cinema{
		ID:               1,
		Name:             "Тестовый фильм",
		Year:             2024,
		TypeNumber:       1, 
		AgeRating:        16,
		Description:      "Длинное описание фильма",
		ShortDescription: "Короткое описание",
		MovieLength:      120,
		Countries: []struct {
			Name string `json:"name"`
		}{
			{Name: "США"},
		},
		Genres: []struct {
			Name string `json:"name"`
		}{
			{Name: "Приключение"},
			{Name: "Хоррор"},
		},
		Poster: &struct {
			URL string `json:"url,omitempty"`
		}{
			URL: "https://example.com/poster.jpg",
		},
		Rating: struct {
			Imdb float32 `json:"imdb"`
			Kp   float32 `json:"kp"`
		}{
			Imdb: 8.5,
			Kp:   7.5,
		},
	}

	filmInfo, picURL, err := movies.FormatMovieInfo(testMovie)

	assert.NoError(t, err)

	assert.Contains(t, filmInfo, "Тестовый фильм")
	assert.Contains(t, filmInfo, "2024")
	assert.Contains(t, filmInfo, "16+")
	assert.Contains(t, filmInfo, "Приключение")
	assert.Contains(t, filmInfo, "Хоррор")
	assert.Contains(t, filmInfo, "Длительность: 2 ч 0 мин")
	assert.Contains(t, filmInfo, "Страны: США")
	assert.Contains(t, filmInfo, "Длинное описание фильма")
	assert.Equal(t, "https://example.com/poster.jpg", picURL)
}
// фильм не найден
func TestFormatMovieInfo_EmptyMovie(t *testing.T) {
	movie := api.Cinema{}

	formattedInfo, imageURL, err := movies.FormatMovieInfo(&movie)

	assert.NoError(t, err)
	assert.NotEmpty(t, formattedInfo)
	assert.Contains(t, formattedInfo, "Фильм не найден")
	assert.Equal(t, "https://habrastorage.org/webt/bh/ex/-z/bhex-zst09dlgq-y2rjespcpp0c.png", imageURL)
}
