package media

import (
	"context"
	"sync/atomic"
)

type Wrk2APIService interface {
	RegisterUser(ctx context.Context, firstName string, lastName string, username string, password string) error
	WriteCastInfo(ctx context.Context, castInfoID int, name string, gender bool, intro string) error
	ComposeReview(ctx context.Context, title string, text string, username string, password string, rating int) error
	RegisterMovie(ctx context.Context, title string, movieID string) error
	WritePlot(ctx context.Context, plotID int, plot string) error
	WriteMovieInfo(ctx context.Context, movieID string, title string, casts []Cast, plotID int, thumbnailIDs []string, photoIDs []string, videoIDs []string, avgRating float64, numRating int) error
}

type Wrk2APIServiceImpl struct {
	userService       UserService
	castInfoService   CastInfoService
	textService       TextService
	plotService       PlotService
	movieIdService    MovieIdService
	movieInfoService  MovieInfoService
	uniqueIdService   UniqueIdService
	requestIDSequence atomic.Int64
}

func NewWrk2APIServiceImpl(ctx context.Context, userService UserService, castInfoService CastInfoService, textService TextService, plotService PlotService, movieIdService MovieIdService, movieInfoService MovieInfoService, uniqueIdService UniqueIdService) (Wrk2APIService, error) {
	return &Wrk2APIServiceImpl{userService: userService, castInfoService: castInfoService, textService: textService, plotService: plotService, movieIdService: movieIdService, movieInfoService: movieInfoService, uniqueIdService: uniqueIdService}, nil
}

func (w *Wrk2APIServiceImpl) nextRequestID() int { return int(w.requestIDSequence.Add(1)) }

func (w *Wrk2APIServiceImpl) RegisterUser(ctx context.Context, firstName string, lastName string, username string, password string) error {
	return w.userService.RegisterUser(ctx, w.nextRequestID(), firstName, lastName, username, password)
}

func (w *Wrk2APIServiceImpl) WriteCastInfo(ctx context.Context, castInfoID int, name string, gender bool, intro string) error {
	return w.castInfoService.WriteCastInfo(ctx, w.nextRequestID(), castInfoID, name, gender, intro)
}

func (w *Wrk2APIServiceImpl) ComposeReview(ctx context.Context, title string, text string, username string, password string, rating int) error {
	reqID := w.nextRequestID()
	if _, err := w.userService.Login(ctx, reqID, username, password); err != nil {
		return err
	}
	if err := w.userService.UploadUserWithUsername(ctx, reqID, username); err != nil {
		return err
	}
	if err := w.textService.UploadText(ctx, reqID, text); err != nil {
		return err
	}
	if err := w.movieIdService.UploadMovieId(ctx, reqID, title, rating); err != nil {
		return err
	}
	return w.uniqueIdService.UploadUniqueId(ctx, int64(reqID))
}

func (w *Wrk2APIServiceImpl) WritePlot(ctx context.Context, plotID int, plot string) error {
	return w.plotService.WritePlot(ctx, w.nextRequestID(), plotID, plot)
}

func (w *Wrk2APIServiceImpl) RegisterMovie(ctx context.Context, title string, movieID string) error {
	return w.movieIdService.RegisterMovieId(ctx, w.nextRequestID(), title, movieID)
}

func (w *Wrk2APIServiceImpl) WriteMovieInfo(ctx context.Context, movieID string, title string, casts []Cast, plotID int, thumbnailIDs []string, photoIDs []string, videoIDs []string, avgRating float64, numRating int) error {
	return w.movieInfoService.WriteMovieInfo(ctx, w.nextRequestID(), movieID, title, casts, plotID, thumbnailIDs, photoIDs, videoIDs, avgRating, numRating)
}
