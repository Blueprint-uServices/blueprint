package workloadgen

import (
	"encoding/json"
	"os"
)

type Cast struct {
	Name      string `json:"name"`
	Gender    int    `json:"gender"`
	Id        int    `json:"id"`
	Biography string `json:"biography"`
}

type CastInfo struct {
	Id        int    `json:"id"`
	CastId    int    `json:"cast_id"`
	Name      string `json:"name"`
	Character string `json:"character"`
}

type Movie struct {
	Title        string     `json:"title"`
	Cast         []CastInfo `json:"cast"`
	Rating       float64    `json:"vote_average"`
	NumRating    int        `json:"vote_count"`
	Id           int        `json:"id"`
	Plot         string     `json:"overview"`
	PosterPath   string     `json:"poster_path"`
	BackdropPath string     `json:"backdrop_path"`
}

func readMovies(filename string) ([]Movie, error) {
	var movies []Movie
	data, err := os.ReadFile(filename)
	if err != nil {
		return movies, err
	}
	if err := json.Unmarshal(data, &movies); err != nil {
		return movies, err
	}
	return movies, nil
}

func readCast(filename string) ([]Cast, error) {
	var casts []Cast
	data, err := os.ReadFile(filename)
	if err != nil {
		return casts, err
	}
	if err := json.Unmarshal(data, &casts); err != nil {
		return casts, err
	}
	return casts, nil
}
