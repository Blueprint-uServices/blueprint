package wiring

import (
	"testing"

	"gitlab.mpi-sws.org/cld/blueprint/plugins/simple"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/workflow"
)

func TestSimpleCache(t *testing.T) {
	spec := newWiringSpec("TestSimpleCache")

	leaf_cache := simple.Cache(spec, "leaf_cache")
	leaf := workflow.Define(spec, "leaf", "TestLeafServiceImplWithCache", leaf_cache)

	app := assertBuildSuccess(t, spec, leaf, leaf_cache)

	assertIR(t, app,
		`TestSimpleCache = BlueprintApplication() {
            leaf.handler.visibility
            leaf_cache.backend.visibility
            leaf_cache = SimpleCache()
            leaf = TestLeafService(leaf_cache)
          }`)
}
func TestSimpleCacheAndServices(t *testing.T) {
	spec := newWiringSpec("TestSimpleCache")

	leaf_cache := simple.Cache(spec, "leaf_cache")
	leaf := workflow.Define(spec, "leaf", "TestLeafServiceImplWithCache", leaf_cache)
	nonleaf := workflow.Define(spec, "nonleaf", "TestNonLeafService", leaf)

	app := assertBuildSuccess(t, spec, leaf, leaf_cache, nonleaf)

	assertIR(t, app,
		`TestSimpleCache = BlueprintApplication() {
            leaf.handler.visibility
            leaf_cache.backend.visibility
            leaf_cache = SimpleCache()
            leaf = TestLeafService(leaf_cache)
            nonleaf.handler.visibility
            nonleaf = TestNonLeafService(leaf)
          }`)
}
