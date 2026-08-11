package workloadgen

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/blueprint-uservices/blueprint/examples/dsb_mm/workflow/media"
	"github.com/blueprint-uservices/blueprint/runtime/core/workload"
)

var outfile = flag.String("outfile", "stats.csv", "Outfile where individual request information will be stored")
var duration = flag.String("duration", "1m", "Duration for which the workload should run")
var tput = flag.Int64("tput", 1000, "Desired throughput/request rate")
var load_data = flag.Bool("load", false, "Load initial data. If selected then only initial data is loaded")
var movies_file = flag.String("movies_file", "movies.json", "Path t movies.json file")
var casts_file = flag.String("casts_file", "casts.json", "Path to casts.json file")

type MediaWorkload interface {
	ImplementsMediaWorkload(ctx context.Context) error
}

type mediaWldGen struct {
	MediaWorkload

	api media.Wrk2APIService
}

func NewMediaWorkload(ctx context.Context, api media.Wrk2APIService) (MediaWorkload, error) {
	w := &mediaWldGen{api: api}
	return w, nil
}

type FnType func() error

func statWrapper(fn FnType) workload.Stat {
	start := time.Now()
	err := fn()
	duration := time.Since(start)
	s := workload.Stat{}
	s.Start = start.UnixNano()
	s.Duration = duration.Nanoseconds()
	s.IsError = (err != nil)
	return s
}

func (w *mediaWldGen) LoadUsers(ctx context.Context) error {
	for i := 0; i < 1000; i++ {
		first_name := fmt.Sprintf("first_name_%d", i)
		last_name := fmt.Sprintf("last_name_%d", i)
		username := fmt.Sprintf("username_%d", i)
		password := fmt.Sprintf("password_%d", i)
		err := w.api.RegisterUser(ctx, first_name, last_name, username, password)
		if err != nil {
			return err
		}
	}
	return nil
}

func (w *mediaWldGen) LoadCast(ctx context.Context, casts []Cast) error {
	for _, cast := range casts {
		var gender bool
		if cast.Gender == 2 {
			gender = true
		} else {
			gender = false
		}
		err := w.api.WriteCastInfo(ctx, cast.Id, cast.Name, gender, cast.Biography)
		if err != nil {
			return err
		}
	}
	return nil
}

func (w *mediaWldGen) LoadMovies(ctx context.Context, movies []Movie) error {
	for _, movie := range movies {
		movie_id := fmt.Sprintf("%v", movie.Id)
		movie_titles = append(movie_titles, movie.Title)
		err := w.api.RegisterMovie(ctx, movie.Title, movie_id)
		if err != nil {
			return err
		}
		var casts []media.Cast
		thumbnails := []string{movie.PosterPath, movie.BackdropPath}
		for _, cinfo := range movie.Cast {
			var cast media.Cast
			cast.CastID = cinfo.Id
			cast.CastInfoID = cinfo.CastId
			cast.Character = cinfo.Character
			casts = append(casts, cast)
		}
		err = w.api.WriteMovieInfo(ctx, movie_id, movie.Title, casts, movie.Id, thumbnails, []string{}, []string{}, movie.Rating, movie.NumRating)
		if err != nil {
			return err
		}

		err = w.api.WritePlot(ctx, movie.Id, movie.Plot)
		if err != nil {
			return err
		}
	}
	return nil
}

func (w *mediaWldGen) LoadData(ctx context.Context) error {
	log.Println("Loading Users")
	err := w.LoadUsers(ctx)
	if err != nil {
		return err
	}

	log.Println("Loading Cast data")
	casts, err := readCast(*casts_file)
	if err != nil {
		return err
	}

	log.Println("Loaded", len(casts), "actors")

	log.Println("Registering Cast data")
	err = w.LoadCast(ctx, casts)
	if err != nil {
		return err
	}

	log.Println("Loading Movie data")
	movies, err := readMovies(*movies_file)
	if err != nil {
		return err
	}

	log.Println("Loaded", len(movies), "movies")

	log.Println("Registering Movie data")
	err = w.LoadMovies(ctx, movies)
	if err != nil {
		return err
	}

	return nil
}

func (w *mediaWldGen) ComposeReviewHandler(ctx context.Context) workload.Stat {
	title, text, username, password, rating := GenReviewHandler()
	return statWrapper(func() error {
		return w.api.ComposeReview(ctx, title, text, username, password, rating)
	})
}

func (w *mediaWldGen) init_movie_data() error {
	if len(movie_titles) != 0 {
		// Load already happened
		return nil
	}

	log.Println("Loading Movie data")
	movies, err := readMovies(*movies_file)
	if err != nil {
		return err
	}

	log.Println("Loaded", len(movies), "movies")

	for _, movie := range movies {
		movie_titles = append(movie_titles, movie.Title)
	}
	return nil
}

func (w *mediaWldGen) Run(ctx context.Context) error {

	if *load_data {
		return w.LoadData(ctx)
	}

	err := w.init_movie_data()
	if err != nil {
		return err
	}

	wrk := workload.NewWorkload()

	wrk.AddAPI("ComposeReviewHandler", w.ComposeReviewHandler, 100)

	engine, err := workload.NewEngine(*outfile, *tput, *duration, wrk)
	if err != nil {
		return err
	}

	engine.RunOpenLoop(ctx)

	return engine.PrintStats()
}

func (w *mediaWldGen) ImplementsMediaWorkload(context.Context) error {
	return nil
}
