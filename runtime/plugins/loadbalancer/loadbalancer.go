package loadbalancer

import (
	"context"
	"math/rand"
)

type LoadBalancer[T any] struct {
	Clients []T
}

func NewLoadBalancer[T any](ctx context.Context, clients []T) *LoadBalancer[T] {
	return &LoadBalancer[T]{Clients: clients}
}

func (this *LoadBalancer[T]) PickClient(ctx context.Context) T {
	// TODO: Support more policies!
	randIndex := rand.Intn(len(this.Clients))
	return this.Clients[randIndex]
}
