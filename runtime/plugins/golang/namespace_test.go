package golang_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/blueprint-uservices/blueprint/runtime/plugins/golang"
	"github.com/stretchr/testify/assert"
)

func TestEmptyNamespace(t *testing.T) {
	b := golang.NewNamespaceBuilder("TestEmptyNamespace")
	n, err := b.Build(context.Background())
	assert.NoError(t, err)

	key := "something"

	var node any
	err = n.Get(key, &node)
	assert.Error(t, err)
}

func TestMissingNode(t *testing.T) {
	b := golang.NewNamespaceBuilder("TestMissingNode")

	key := "something2"

	b.Required(key, "something required")
	_, err := b.Build(context.Background())
	assert.Error(t, err)
}

func TestExistingNode(t *testing.T) {
	b := golang.NewNamespaceBuilder("TestExistingNode")

	key := "something3"

	b.Set(key, "good")
	n, err := b.Build(context.Background())
	assert.NoError(t, err)

	var node string
	err = n.Get(key, &node)
	assert.NoError(t, err)
	assert.Equal(t, "good", node)
}

func TestMissingNodeWithParent(t *testing.T) {
	b1 := golang.NewNamespaceBuilder("TestMissingNodeWithParent-Parent")
	n1, err := b1.Build(context.Background())
	assert.NoError(t, err)

	b2 := golang.NewNamespaceBuilder("TestMissingNodeWithParent-Child")

	key := "something4"

	b2.Required(key, "something required")
	_, err = b2.BuildWithParent(n1)
	assert.Error(t, err)
}

func TestParentNode(t *testing.T) {
	b1 := golang.NewNamespaceBuilder("TestParentNode-Parent")

	key := "something5"

	b1.Set(key, "good")
	n1, err := b1.Build(context.Background())
	assert.NoError(t, err)

	b2 := golang.NewNamespaceBuilder("TestParentNode-Child")
	b2.Required(key, "something required")
	n2, err := b2.BuildWithParent(n1)
	assert.NoError(t, err)

	var node string
	err = n2.Get(key, &node)
	assert.NoError(t, err)
	assert.Equal(t, "good", node)
}

func TestBuildOnce(t *testing.T) {
	b := golang.NewNamespaceBuilder("TestBuildOnce")
	count := 0

	key := "something6"

	b.Define(key, func(n *golang.Namespace) (any, error) {
		count = count + 1
		return "hello", nil
	})
	n, err := b.Build(context.Background())
	assert.NoError(t, err)

	var node string
	err = n.Get(key, &node)
	assert.NoError(t, err)
	assert.Equal(t, "hello", node)
	assert.Equal(t, 1, count)

	err = n.Get(key, &node)
	assert.NoError(t, err)
	assert.Equal(t, "hello", node)
	assert.Equal(t, 1, count)
}

func TestBuildInParentNamespace(t *testing.T) {
	pb := golang.NewNamespaceBuilder("TestBuildInParentNamespace-Parent")
	count := 0

	key := "something7"

	pb.Define(key, func(n *golang.Namespace) (any, error) {
		count = count + 1
		return "hello", nil
	})
	p, err := pb.Build(context.Background())
	assert.NoError(t, err)

	cb := golang.NewNamespaceBuilder("TestBuildInParentNamespace-Child")
	cb.Required(key, "something required")
	c, err := cb.BuildWithParent(p)
	assert.NoError(t, err)

	var node string
	err = c.Get(key, &node)
	assert.NoError(t, err)
	assert.Equal(t, "hello", node)
	assert.Equal(t, 1, count)

	err = p.Get(key, &node)
	assert.NoError(t, err)
	assert.Equal(t, "hello", node)
	assert.Equal(t, 1, count)
}

func TestBuildError(t *testing.T) {
	b := golang.NewNamespaceBuilder("TestBuildError")
	count := 0

	key := "something8"

	b.Define(key, func(n *golang.Namespace) (any, error) {
		count = count + 1
		return nil, fmt.Errorf("uhoh")
	})
	n, err := b.Build(context.Background())
	assert.NoError(t, err)

	var node string
	err = n.Get(key, &node)
	assert.Error(t, err)
	assert.Equal(t, 1, count)
	err = n.Get(key, &node)
	assert.Error(t, err)
	assert.Equal(t, 2, count)
}

func TestGetInBuildFunc(t *testing.T) {
	b := golang.NewNamespaceBuilder("TestBuildError")
	count := 0

	key := "something9"

	b.Define(key, func(n *golang.Namespace) (any, error) {
		count = count + 1
		return "hello", nil
	})
	b.Define("somethingelse", func(n *golang.Namespace) (any, error) {
		var something string
		err := n.Get(key, &something)
		return something + " world", err
	})
	n, err := b.Build(context.Background())
	assert.NoError(t, err)

	var node string
	err = n.Get("somethingelse", &node)
	assert.NoError(t, err)
	assert.Equal(t, "hello world", node)
	assert.Equal(t, 1, count)
}

func TestInstantiate(t *testing.T) {
	b := golang.NewNamespaceBuilder("TestBuildError")
	count := 0

	key := "something10"

	b.Define(key, func(n *golang.Namespace) (any, error) {
		count = count + 1
		return "hello", nil
	})
	b.Instantiate(key)
	_, err := b.Build(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, 1, count)
}

type runtester struct {
	done bool
	golang.Runnable
}

func (r *runtester) Run(ctx context.Context) error {
	select {
	case <-ctx.Done():
		{
			r.done = true
		}
	}
	return nil
}

func TestRun(t *testing.T) {
	b := golang.NewNamespaceBuilder("TestBuildError")
	tester := &runtester{}

	key := "something11"

	b.Define(key, func(n *golang.Namespace) (any, error) {
		return tester, nil
	})
	n, err := b.Build(context.Background())
	assert.NoError(t, err)

	assert.False(t, tester.done)

	var tester2 *runtester
	err = n.Get(key, &tester2)
	assert.False(t, tester.done)
	assert.Equal(t, tester, tester2)

	time.Sleep(100 * time.Millisecond)

	assert.False(t, tester.done)
	n.Shutdown(true)
	assert.True(t, tester.done)
}

func TestNestedRun(t *testing.T) {
	pb := golang.NewNamespaceBuilder("TestBuildInParentNamespace-Parent")
	tester1 := &runtester{}

	key := "something12"

	pb.Define(key, func(n *golang.Namespace) (any, error) {
		return tester1, nil
	})
	p, err := pb.Build(context.Background())
	assert.NoError(t, err)

	cb := golang.NewNamespaceBuilder("TestBuildInParentNamespace-Child")
	tester2 := &runtester{}
	cb.Define("somethingelse", func(n *golang.Namespace) (any, error) {
		var node any
		err := n.Get(key, &node)
		return tester2, err
	})
	c, err := cb.BuildWithParent(p)
	assert.NoError(t, err)

	var tester3 *runtester
	err = c.Get("somethingelse", &tester3)
	assert.NoError(t, err)
	assert.Equal(t, tester2, tester3)

	assert.False(t, tester1.done)
	assert.False(t, tester2.done)

	time.Sleep(100 * time.Millisecond)

	assert.False(t, tester1.done)
	assert.False(t, tester2.done)
	p.Shutdown(true)
	assert.True(t, tester1.done)
	assert.True(t, tester2.done)
}
