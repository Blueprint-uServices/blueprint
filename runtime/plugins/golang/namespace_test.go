package golang_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gitlab.mpi-sws.org/cld/blueprint/runtime/plugins/golang"
)

func TestEmptyNamespace(t *testing.T) {
	b := golang.NewNamespaceBuilder("TestEmptyNamespace")
	n, err := b.Build(context.Background())
	assert.NoError(t, err)

	var node any
	err = n.Get("something", &node)
	assert.Error(t, err)
}

func TestMissingNode(t *testing.T) {
	b := golang.NewNamespaceBuilder("TestMissingNode")
	b.Required("something", "something required")
	_, err := b.Build(context.Background())
	assert.Error(t, err)
}

func TestExistingNode(t *testing.T) {
	b := golang.NewNamespaceBuilder("TestExistingNode")
	b.Set("something", "good")
	n, err := b.Build(context.Background())
	assert.NoError(t, err)

	var node string
	err = n.Get("something", &node)
	assert.NoError(t, err)
	assert.Equal(t, "good", node)
}

func TestMissingNodeWithParent(t *testing.T) {
	b1 := golang.NewNamespaceBuilder("TestMissingNodeWithParent-Parent")
	n1, err := b1.Build(context.Background())
	assert.NoError(t, err)

	b2 := golang.NewNamespaceBuilder("TestMissingNodeWithParent-Child")
	b2.Required("something", "something required")
	_, err = b2.BuildWithParent(n1)
	assert.Error(t, err)
}

func TestParentNode(t *testing.T) {
	b1 := golang.NewNamespaceBuilder("TestParentNode-Parent")
	b1.Set("something", "good")
	n1, err := b1.Build(context.Background())
	assert.NoError(t, err)

	b2 := golang.NewNamespaceBuilder("TestParentNode-Child")
	b2.Required("something", "something required")
	n2, err := b2.BuildWithParent(n1)
	assert.NoError(t, err)

	var node string
	err = n2.Get("something", &node)
	assert.NoError(t, err)
	assert.Equal(t, "good", node)
}

func TestBuildOnce(t *testing.T) {
	b := golang.NewNamespaceBuilder("TestBuildOnce")
	count := 0
	b.Define("something", func(n *golang.Namespace) (any, error) {
		count = count + 1
		return "hello", nil
	})
	n, err := b.Build(context.Background())
	assert.NoError(t, err)

	var node string
	err = n.Get("something", &node)
	assert.NoError(t, err)
	assert.Equal(t, "hello", node)
	assert.Equal(t, 1, count)

	err = n.Get("something", &node)
	assert.NoError(t, err)
	assert.Equal(t, "hello", node)
	assert.Equal(t, 1, count)
}

func TestBuildInParentNamespace(t *testing.T) {
	pb := golang.NewNamespaceBuilder("TestBuildInParentNamespace-Parent")
	count := 0
	pb.Define("something", func(n *golang.Namespace) (any, error) {
		count = count + 1
		return "hello", nil
	})
	p, err := pb.Build(context.Background())
	assert.NoError(t, err)

	cb := golang.NewNamespaceBuilder("TestBuildInParentNamespace-Child")
	cb.Required("something", "something required")
	c, err := cb.BuildWithParent(p)
	assert.NoError(t, err)

	var node string
	err = c.Get("something", &node)
	assert.NoError(t, err)
	assert.Equal(t, "hello", node)
	assert.Equal(t, 1, count)

	err = p.Get("something", &node)
	assert.NoError(t, err)
	assert.Equal(t, "hello", node)
	assert.Equal(t, 1, count)
}

func TestBuildError(t *testing.T) {
	b := golang.NewNamespaceBuilder("TestBuildError")
	count := 0
	b.Define("something", func(n *golang.Namespace) (any, error) {
		count = count + 1
		return nil, fmt.Errorf("uhoh")
	})
	n, err := b.Build(context.Background())
	assert.NoError(t, err)

	var node string
	err = n.Get("something", &node)
	assert.Error(t, err)
	assert.Equal(t, 1, count)
	err = n.Get("something", &node)
	assert.Error(t, err)
	assert.Equal(t, 2, count)
}

func TestGetInBuildFunc(t *testing.T) {
	b := golang.NewNamespaceBuilder("TestBuildError")
	count := 0
	b.Define("something", func(n *golang.Namespace) (any, error) {
		count = count + 1
		return "hello", nil
	})
	b.Define("somethingelse", func(n *golang.Namespace) (any, error) {
		var something string
		err := n.Get("something", &something)
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
	b.Define("something", func(n *golang.Namespace) (any, error) {
		count = count + 1
		return "hello", nil
	})
	b.Instantiate("something")
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
	b.Define("something", func(n *golang.Namespace) (any, error) {
		return tester, nil
	})
	n, err := b.Build(context.Background())
	assert.NoError(t, err)

	assert.False(t, tester.done)

	var tester2 *runtester
	err = n.Get("something", &tester2)
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
	pb.Define("something", func(n *golang.Namespace) (any, error) {
		return tester1, nil
	})
	p, err := pb.Build(context.Background())
	assert.NoError(t, err)

	cb := golang.NewNamespaceBuilder("TestBuildInParentNamespace-Child")
	tester2 := &runtester{}
	cb.Define("somethingelse", func(n *golang.Namespace) (any, error) {
		var node any
		err := n.Get("something", &node)
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
