package boutique

import (
	"context"
	"fmt"
	"math"
	"sort"

	"github.com/blueprint-uservices/blueprint/runtime/core/backend"
)

type CurrencyService interface {
	GetSupportedCurrencies(ctx context.Context) ([]string, error)
	Convert(ctx context.Context, from Money, toCode string) (Money, error)
}

type CurrencyServiceImpl struct {
	conversionMap map[string]float64
	logger        backend.Logger
}

func NewCurrencyServiceImpl(ctx context.Context) (CurrencyService, error) {
	return &CurrencyServiceImpl{conversionMap: currencyConversions(), logger: backend.GetLogger()}, nil
}

func (s *CurrencyServiceImpl) GetSupportedCurrencies(ctx context.Context) ([]string, error) {
	ctx, _ = s.logger.Info(ctx, "Getting supported currencies")
	currencies := make([]string, 0, len(s.conversionMap))
	for code := range s.conversionMap {
		currencies = append(currencies, code)
	}
	sort.Strings(currencies)
	return currencies, nil
}

func (s *CurrencyServiceImpl) Convert(ctx context.Context, from Money, toCode string) (Money, error) {
	fromRate, ok := s.conversionMap[from.CurrencyCode]
	if !ok {
		return Money{}, fmt.Errorf("unsupported currency code %q", from.CurrencyCode)
	}
	toRate, ok := s.conversionMap[toCode]
	if !ok {
		return Money{}, fmt.Errorf("unsupported currency code %q", toCode)
	}
	euros := carry(float64(from.Units)/fromRate, float64(from.Nanos)/fromRate)
	to := carry(float64(euros.Units)*toRate, float64(euros.Nanos)*toRate)
	to.CurrencyCode = toCode
	return to, nil
}

func carry(units, nanos float64) Money {
	const fractionSize = 1_000_000_000
	nanos += math.Mod(units, 1) * fractionSize
	units = math.Floor(units) + math.Floor(nanos/fractionSize)
	nanos = math.Mod(nanos, fractionSize)
	return Money{Units: int64(units), Nanos: int32(nanos)}
}

func currencyConversions() map[string]float64 {
	return map[string]float64{
		"EUR": 1, "USD": 1.1305, "JPY": 126.40, "BGN": 1.9558,
		"CZK": 25.592, "DKK": 7.4609, "GBP": 0.85970, "HUF": 315.51,
		"PLN": 4.2996, "RON": 4.7463, "SEK": 10.5375, "CHF": 1.1360,
		"ISK": 136.80, "NOK": 9.8040, "HRK": 7.4210, "RUB": 74.4208,
		"TRY": 6.1247, "AUD": 1.6072, "BRL": 4.2682, "CAD": 1.5128,
		"CNY": 7.5857, "HKD": 8.8743, "IDR": 15999.40, "ILS": 4.0875,
		"INR": 79.4320, "KRW": 1275.05, "MXN": 21.7999, "MYR": 4.6289,
		"NZD": 1.6679, "PHP": 59.083, "SGD": 1.5349, "THB": 36.012,
		"ZAR": 16.0583,
	}
}
