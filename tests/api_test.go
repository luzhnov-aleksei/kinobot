package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/luzhnov-aleksei/kinobot/api"
	"github.com/stretchr/testify/assert"
)

func mockSuccessfulResponse() *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1.4/movie/search", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"docs": [
				{
					"id": 1,
					"name": "Test Movie",
					"year": 2024,
					"typeNumber": 1,
					"ageRating": 16,
					"description": "Test description",
					"shortDescription": "Short desc",
					"movieLength": 120,
					"countries": [{"name": "USA"}],
					"genres": [{"name": "Action"}, {"name": "Adventure"}],
					"poster": {"url": "https://example.com/poster.jpg"},
					"rating": {"imdb": 8.5, "kp": 7.5}
				}
			]
		}`))
	})
	return httptest.NewServer(mux)
}

func TestRequestMovies_Success(t *testing.T) {
	server := mockSuccessfulResponse()
	defer server.Close()

	movies, err := api.RequestMovies(server.URL+"/v1.4/movie/search", "Test Movie")
	assert.NoError(t, err)
	assert.Len(t, movies, 1)
	assert.Equal(t, uint32(1), movies[0].ID)
	assert.Equal(t, "Test Movie", movies[0].Name)
	assert.Equal(t, "https://example.com/poster.jpg", movies[0].Poster.URL)
}

func TestRequestMovies_Error(t *testing.T) {
	server := httptest.NewServer(http.NotFoundHandler())
	defer server.Close()

	movies, err := api.RequestMovies(server.URL+"/v1.4/movie/search", "Test Movie")
	assert.Error(t, err)
	assert.Nil(t, movies)
}
