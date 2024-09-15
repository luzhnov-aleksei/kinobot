package api

import (
	"testing"

	"github.com/luzhnov-aleksei/kinobot/api"
	"github.com/luzhnov-aleksei/kinobot/movies"
	"github.com/stretchr/testify/assert"
)


func TestUserMovieSelections(t *testing.T) {
	movies.UserMovieSelections = make(map[int64][]api.Cinema)

	testMovies := []api.Cinema{
		{ID: 1, Name: "Фильм 1"},
		{ID: 2, Name: "Фильм 2"},
	}

	userID := int64(123)
	movies.UserMovieSelections[userID] = testMovies

	savedMovies, exists := movies.UserMovieSelections[userID]
	assert.True(t, exists)
	assert.Equal(t, 2, len(savedMovies))
	assert.Equal(t, "Фильм 1", savedMovies[0].Name)
	assert.Equal(t, "Фильм 2", savedMovies[1].Name)
}