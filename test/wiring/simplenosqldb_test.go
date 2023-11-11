package wiring

import (
	"testing"

	"gitlab.mpi-sws.org/cld/blueprint/plugins/simplecache"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/simplenosqldb"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/workflow"
)

func TestSimpleNoSQLDB(t *testing.T) {
	spec := newWiringSpec("TestSimpleNoSQLDB")

	leaf_cache := simplecache.Define(spec, "leaf_cache")
	leaf_db := simplenosqldb.Define(spec, "leaf_db")
	leaf := workflow.Define(spec, "leaf", "TestLeafServiceImplWithDB", leaf_cache, leaf_db)
	nonleaf := workflow.Define(spec, "nonleaf", "TestNonLeafService", leaf)

	app := assertBuildSuccess(t, spec, leaf, leaf_db, nonleaf)

	assertIR(t, app,
		`TestSimpleNoSQLDB = BlueprintApplication() {
			leaf.handler.visibility
			leaf_cache.backend.visibility
			leaf_cache = SimpleCache()
			leaf_db.backend.visibility
			leaf_db = SimpleNoSQLDB()
			leaf = TestLeafService(leaf_cache, leaf_db)
			nonleaf.handler.visibility
			nonleaf = TestNonLeafService(leaf)
		  }`)
}
