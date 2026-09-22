package boutique

import (
	"context"
	"errors"
	"strings"

	"github.com/blueprint-uservices/blueprint/runtime/core/backend"
)

var ErrProductNotFound = errors.New("product not found")

type ProductCatalogService interface {
	ListProducts(ctx context.Context) ([]Product, error)
	GetProduct(ctx context.Context, productID string) (Product, error)
	SearchProducts(ctx context.Context, query string) ([]Product, error)
}

type ProductCatalogServiceImpl struct {
	products []Product
	logger   backend.Logger
}

func NewProductCatalogServiceImpl(ctx context.Context) (ProductCatalogService, error) {
	return &ProductCatalogServiceImpl{products: productCatalog(), logger: backend.GetLogger()}, nil
}

func (s *ProductCatalogServiceImpl) ListProducts(ctx context.Context) ([]Product, error) {
	return append([]Product(nil), s.products...), nil
}

func (s *ProductCatalogServiceImpl) GetProduct(ctx context.Context, productID string) (Product, error) {
	for _, product := range s.products {
		if product.ID == productID {
			return product, nil
		}
	}
	return Product{}, ErrProductNotFound
}

func (s *ProductCatalogServiceImpl) SearchProducts(ctx context.Context, query string) ([]Product, error) {
	query = strings.ToLower(query)
	var products []Product
	for _, product := range s.products {
		if strings.Contains(strings.ToLower(product.Name), query) || strings.Contains(strings.ToLower(product.Description), query) {
			products = append(products, product)
		}
	}
	return products, nil
}

func productCatalog() []Product {
	usd := func(units int64, nanos int32) Money { return Money{CurrencyCode: "USD", Units: units, Nanos: nanos} }
	return []Product{
		{ID: "OLJCESPC7Z", Name: "Sunglasses", Description: "Add a modern touch to your outfits with these sleek aviator sunglasses.", Picture: "/static/img/products/sunglasses.jpg", PriceUSD: usd(19, 990000000), Categories: []string{"accessories"}},
		{ID: "66VCHSJNUP", Name: "Tank Top", Description: "Perfectly cropped cotton tank, with a scooped neckline.", Picture: "/static/img/products/tank-top.jpg", PriceUSD: usd(18, 990000000), Categories: []string{"clothing", "tops"}},
		{ID: "1YMWWN1N4O", Name: "Watch", Description: "This gold-tone stainless steel watch will work with most of your outfits.", Picture: "/static/img/products/watch.jpg", PriceUSD: usd(109, 990000000), Categories: []string{"accessories"}},
		{ID: "L9ECAV7KIM", Name: "Loafers", Description: "A neat addition to your summer wardrobe.", Picture: "/static/img/products/loafers.jpg", PriceUSD: usd(89, 990000000), Categories: []string{"footwear"}},
		{ID: "2ZYFJ3GM2N", Name: "Hairdryer", Description: "This lightweight hairdryer has 3 heat and speed settings. It's perfect for travel.", Picture: "/static/img/products/hairdryer.jpg", PriceUSD: usd(24, 990000000), Categories: []string{"hair", "beauty"}},
		{ID: "0PUK6V6EV0", Name: "Candle Holder", Description: "This small but intricate candle holder is an excellent gift.", Picture: "/static/img/products/candle-holder.jpg", PriceUSD: usd(18, 990000000), Categories: []string{"decor", "home"}},
		{ID: "LS4PSXUNUM", Name: "Salt & Pepper Shakers", Description: "Add some flavor to your kitchen.", Picture: "/static/img/products/salt-and-pepper-shakers.jpg", PriceUSD: usd(18, 490000000), Categories: []string{"kitchen"}},
		{ID: "9SIQT8TOJO", Name: "Bamboo Glass Jar", Description: "This bamboo glass jar can hold 57 oz (1.7 l) and is perfect for any kitchen.", Picture: "/static/img/products/bamboo-glass-jar.jpg", PriceUSD: usd(5, 490000000), Categories: []string{"kitchen"}},
		{ID: "6E92ZMYYFZ", Name: "Mug", Description: "A simple mug with a mustard interior.", Picture: "/static/img/products/mug.jpg", PriceUSD: usd(8, 990000000), Categories: []string{"kitchen"}},
	}
}
