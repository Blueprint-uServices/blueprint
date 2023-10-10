package wiring

import (
	"testing"

	"gitlab.mpi-sws.org/cld/blueprint/plugins/simplecache"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/workflow"
)

func TestSimpleCache(t *testing.T) {
	wiring := newWiringSpec("TestSimpleCache")

	leaf_cache := simplecache.Define(wiring, "leaf_cache")
	leaf := workflow.Define(wiring, "leaf", "TestLeafServiceImplWithCache", leaf_cache)

	app := assertBuildSuccess(t, wiring, leaf, leaf_cache)

	assertIR(t, app,
		`TestSimpleCache = BlueprintApplication() {
            leaf.handler.visibility
            leaf_cache.backend.visibility
            leaf_cache = SimpleCache()
            leaf = TestLeafService(leaf_cache)
          }`)
}
func TestSimpleCacheAndServices(t *testing.T) {
	wiring := newWiringSpec("TestSimpleCache")

	leaf_cache := simplecache.Define(wiring, "leaf_cache")
	leaf := workflow.Define(wiring, "leaf", "TestLeafServiceImplWithCache", leaf_cache)
	nonleaf := workflow.Define(wiring, "nonleaf", "TestNonLeafService", leaf)

	app := assertBuildSuccess(t, wiring, leaf, leaf_cache, nonleaf)

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
