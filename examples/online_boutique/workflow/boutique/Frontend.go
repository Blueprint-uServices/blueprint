package boutique

import (
	"context"
	"fmt"

	"github.com/blueprint-uservices/blueprint/runtime/core/backend"
)

type ProductView struct {
	Item  Product
	Price Money
}

type HomeView struct {
	Products     []ProductView
	Currencies   []string
	CartSize     int32
	Ad           Ad
	UserCurrency string
}

type ProductPageView struct {
	Product         ProductView
	Recommendations []Product
	Currencies      []string
	CartSize        int32
	Ad              Ad
}

type CartItemView struct {
	Item     Product
	Quantity int32
	Price    Money
}

type CartView struct {
	Items           []CartItemView
	Recommendations []Product
	ShippingCost    Money
	TotalCost       Money
}

type Frontend interface {
	Home(ctx context.Context, userID, currency string) (HomeView, error)
	GetProduct(ctx context.Context, userID, productID, currency string) (ProductPageView, error)
	AddToCart(ctx context.Context, userID, productID string, quantity int32) error
	GetCart(ctx context.Context, userID, currency string) (CartView, error)
	EmptyCart(ctx context.Context, userID string) error
	PlaceOrder(ctx context.Context, req PlaceOrderRequest) (Order, error)
}

type FrontendImpl struct {
	catalogService        ProductCatalogService
	currencyService       CurrencyService
	cartService           CartService
	recommendationService RecommendationService
	checkoutService       CheckoutService
	shippingService       ShippingService
	adService             AdService
	logger                backend.Logger
}

func NewFrontendImpl(
	ctx context.Context,
	catalogService ProductCatalogService,
	currencyService CurrencyService,
	cartService CartService,
	recommendationService RecommendationService,
	checkoutService CheckoutService,
	shippingService ShippingService,
	adService AdService,
) (Frontend, error) {
	return &FrontendImpl{
		catalogService: catalogService, currencyService: currencyService, cartService: cartService,
		recommendationService: recommendationService, checkoutService: checkoutService,
		shippingService: shippingService, adService: adService, logger: backend.GetLogger(),
	}, nil
}

func (f *FrontendImpl) Home(ctx context.Context, userID, currency string) (HomeView, error) {
	products, err := f.catalogService.ListProducts(ctx)
	if err != nil {
		return HomeView{}, err
	}
	views := make([]ProductView, len(products))
	for i, product := range products {
		price, err := f.currencyService.Convert(ctx, product.PriceUSD, currency)
		if err != nil {
			return HomeView{}, err
		}
		views[i] = ProductView{Item: product, Price: price}
	}
	currencies, err := f.currencyService.GetSupportedCurrencies(ctx)
	if err != nil {
		return HomeView{}, err
	}
	cart, err := f.cartService.GetCart(ctx, userID)
	if err != nil {
		return HomeView{}, err
	}
	ads, err := f.adService.GetAds(ctx, nil)
	if err != nil {
		return HomeView{}, err
	}
	return HomeView{Products: views, Currencies: currencies, CartSize: cartSize(cart), Ad: firstAd(ads), UserCurrency: currency}, nil
}

func (f *FrontendImpl) GetProduct(ctx context.Context, userID, productID, currency string) (ProductPageView, error) {
	product, err := f.catalogService.GetProduct(ctx, productID)
	if err != nil {
		return ProductPageView{}, err
	}
	price, err := f.currencyService.Convert(ctx, product.PriceUSD, currency)
	if err != nil {
		return ProductPageView{}, err
	}
	recommendations, err := f.recommendedProducts(ctx, userID, []string{productID})
	if err != nil {
		return ProductPageView{}, err
	}
	currencies, err := f.currencyService.GetSupportedCurrencies(ctx)
	if err != nil {
		return ProductPageView{}, err
	}
	cart, err := f.cartService.GetCart(ctx, userID)
	if err != nil {
		return ProductPageView{}, err
	}
	ads, err := f.adService.GetAds(ctx, product.Categories)
	if err != nil {
		return ProductPageView{}, err
	}
	return ProductPageView{
		Product: ProductView{Item: product, Price: price}, Recommendations: recommendations,
		Currencies: currencies, CartSize: cartSize(cart), Ad: firstAd(ads),
	}, nil
}

func (f *FrontendImpl) AddToCart(ctx context.Context, userID, productID string, quantity int32) error {
	product, err := f.catalogService.GetProduct(ctx, productID)
	if err != nil {
		return err
	}
	return f.cartService.AddItem(ctx, userID, CartItem{ProductID: product.ID, Quantity: quantity})
}

func (f *FrontendImpl) GetCart(ctx context.Context, userID, currency string) (CartView, error) {
	cart, err := f.cartService.GetCart(ctx, userID)
	if err != nil {
		return CartView{}, err
	}
	shippingUSD, err := f.shippingService.GetQuote(ctx, Address{}, cart)
	if err != nil {
		return CartView{}, err
	}
	shipping, err := f.currencyService.Convert(ctx, shippingUSD, currency)
	if err != nil {
		return CartView{}, err
	}
	view := CartView{Items: make([]CartItemView, len(cart)), ShippingCost: shipping, TotalCost: Money{CurrencyCode: currency}}
	for i, item := range cart {
		product, err := f.catalogService.GetProduct(ctx, item.ProductID)
		if err != nil {
			return CartView{}, err
		}
		unitPrice, err := f.currencyService.Convert(ctx, product.PriceUSD, currency)
		if err != nil {
			return CartView{}, err
		}
		price := MultiplySlow(unitPrice, uint32(item.Quantity))
		view.Items[i] = CartItemView{Item: product, Quantity: item.Quantity, Price: price}
		view.TotalCost, err = Sum(view.TotalCost, price)
		if err != nil {
			return CartView{}, err
		}
	}
	view.TotalCost, err = Sum(view.TotalCost, shipping)
	if err != nil {
		return CartView{}, err
	}
	view.Recommendations, err = f.recommendedProducts(ctx, userID, cartIDs(cart))
	return view, err
}

func (f *FrontendImpl) EmptyCart(ctx context.Context, userID string) error {
	return f.cartService.EmptyCart(ctx, userID)
}

func (f *FrontendImpl) PlaceOrder(ctx context.Context, req PlaceOrderRequest) (Order, error) {
	order, err := f.checkoutService.PlaceOrder(ctx, req)
	if err != nil {
		return Order{}, fmt.Errorf("could not place order: %w", err)
	}
	return order, nil
}

func (f *FrontendImpl) recommendedProducts(ctx context.Context, userID string, excluded []string) ([]Product, error) {
	ids, err := f.recommendationService.ListRecommendations(ctx, userID, excluded)
	if err != nil {
		return nil, err
	}
	products := make([]Product, len(ids))
	for i, id := range ids {
		products[i], err = f.catalogService.GetProduct(ctx, id)
		if err != nil {
			return nil, err
		}
	}
	return products, nil
}

func cartSize(items []CartItem) int32 {
	var size int32
	for _, item := range items {
		size += item.Quantity
	}
	return size
}

func cartIDs(items []CartItem) []string {
	ids := make([]string, len(items))
	for i, item := range items {
		ids[i] = item.ProductID
	}
	return ids
}

func firstAd(ads []Ad) Ad {
	if len(ads) == 0 {
		return Ad{}
	}
	return ads[0]
}
