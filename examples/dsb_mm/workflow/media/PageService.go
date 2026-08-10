package media

import "context"

type PageService interface {
	ReadPage(ctx context.Context, reqID int, movieID string, reviewStart int, reviewEnd int) (Page, error)
}

type PageServiceImpl struct {
	movieInfoService   MovieInfoService
	movieReviewService MovieReviewService
	castInfoService    CastInfoService
	plotService        PlotService
}

func NewPageServiceImpl(movieInfoService MovieInfoService, movieReviewService MovieReviewService, castInfoService CastInfoService, plotService PlotService) (PageService, error) {
	return &PageServiceImpl{movieInfoService: movieInfoService, movieReviewService: movieReviewService, castInfoService: castInfoService, plotService: plotService}, nil
}

func (p *PageServiceImpl) ReadPage(ctx context.Context, reqID int, movieID string, reviewStart int, reviewEnd int) (Page, error) {
	var page Page
	info, err := p.movieInfoService.ReadMovieInfo(ctx, reqID, movieID)
	if err != nil {
		return page, err
	}
	page.MovieInfo = info
	page.Reviews, err = p.movieReviewService.ReadMovieReviews(ctx, reqID, movieID, reviewStart, reviewEnd)
	if err != nil {
		return page, err
	}
	castInfoIDs := make([]int, 0, len(info.Casts))
	for _, cast := range info.Casts {
		castInfoIDs = append(castInfoIDs, cast.CastInfoID)
	}
	page.CastInfo, err = p.castInfoService.ReadCastInfo(ctx, reqID, castInfoIDs)
	if err != nil {
		return page, err
	}
	page.Plot, err = p.plotService.ReadPlot(ctx, reqID, info.PlotID)
	return page, err
}
