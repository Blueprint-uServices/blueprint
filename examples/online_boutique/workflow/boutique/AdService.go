package boutique

import (
	"context"
	"math/rand"

	"github.com/blueprint-uservices/blueprint/runtime/core/backend"
)

const maxAdsToServe = 2

type AdService interface {
	GetAds(ctx context.Context, keywords []string) ([]Ad, error)
}

type AdServiceImpl struct {
	ads    map[string]Ad
	logger backend.Logger
}

func NewAdServiceImpl(ctx context.Context) (AdService, error) {
	s := &AdServiceImpl{ads: createAdsMap(), logger: backend.GetLogger()}
	ctx, _ = s.logger.Info(ctx, "Ad Service started")
	return s, nil
}

func (s *AdServiceImpl) GetAds(ctx context.Context, keywords []string) ([]Ad, error) {
	ctx, _ = s.logger.Info(ctx, "received ad request: keywords=%v", keywords)
	ads := make([]Ad, 0, len(keywords))
	for _, keyword := range keywords {
		if ad, ok := s.ads[keyword]; ok {
			ads = append(ads, ad)
		}
	}
	if len(ads) != 0 {
		return ads, nil
	}
	values := make([]Ad, 0, len(s.ads))
	for _, ad := range s.ads {
		values = append(values, ad)
	}
	ads = make([]Ad, maxAdsToServe)
	for i := range ads {
		ads[i] = values[rand.Intn(len(values))]
	}
	return ads, nil
}

func createAdsMap() map[string]Ad {
	return map[string]Ad{
		"hair":        {RedirectURL: "/product/2ZYFJ3GM2N", Text: "Hairdryer for sale. 50% off."},
		"clothing":    {RedirectURL: "/product/66VCHSJNUP", Text: "Tank top for sale. 20% off."},
		"accessories": {RedirectURL: "/product/1YMWWN1N4O", Text: "Watch for sale. Buy one, get second kit for free"},
		"footwear":    {RedirectURL: "/product/L9ECAV7KIM", Text: "Loafers for sale. Buy one, get second one for free"},
		"decor":       {RedirectURL: "/product/0PUK6V6EV0", Text: "Candle holder for sale. 30% off."},
		"kitchen":     {RedirectURL: "/product/9SIQT8TOJO", Text: "Bamboo glass jar for sale. 10% off."},
	}
}
