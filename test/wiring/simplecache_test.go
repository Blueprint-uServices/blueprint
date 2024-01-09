package wiring

import (
	"testing"

	"github.com/blueprint-uservices/blueprint/plugins/simple"
	"github.com/blueprint-uservices/blueprint/plugins/workflow"
)

func TestSimpleCache(t *testing.T) {
	spec := newWiringSpec("TestSimpleCache")

	leaf_cache := simple.Cache(spec, "leaf_cache")
	leaf := workflow.Service(spec, "leaf", "TestLeafServiceImplWithCache", leaf_cache)

	app := assertBuildSuccess(t, spec, leaf, leaf_cache)

	assertIR(t, app,
		`TestSimpleCache = BlueprintApplication() {
			leaf = TestLeafService(leaf_cache)
			leaf.client = leaf
			leaf.handler.visibility
			leaf_cache = SimpleCache()
			leaf_cache.backend.visibility
          }`)
}
func TestSimpleCacheAndServices(t *testing.T) {
	spec := newWiringSpec("TestSimpleCacheAndServices")

	leaf_cache := simple.Cache(spec, "leaf_cache")
	leaf := workflow.Service(spec, "leaf", "TestLeafServiceImplWithCache", leaf_cache)
	nonleaf := workflow.Service(spec, "nonleaf", "TestNonLeafService", leaf)

	app := assertBuildSuccess(t, spec, leaf, leaf_cache, nonleaf)

	assertIR(t, app,
		`TestSimpleCacheAndServices = BlueprintApplication() {
			leaf = TestLeafService(leaf_cache)
			leaf.client = leaf
			leaf.handler.visibility
			leaf_cache = SimpleCache()
			leaf_cache.backend.visibility
			nonleaf = TestNonLeafService(leaf.client)
			nonleaf.client = nonleaf
			nonleaf.handler.visibility
          }`)
}
