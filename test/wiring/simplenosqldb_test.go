package wiring

import (
	"testing"

	"gitlab.mpi-sws.org/cld/blueprint/plugins/simplecache"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/simplenosqldb"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/workflow"
)

func TestSimpleNoSQLDB(t *testing.T) {
	wiring := newWiringSpec("TestSimpleNoSQLDB")

	leaf_cache := simplecache.Define(wiring, "leaf_cache")
	leaf_db := simplenosqldb.Define(wiring, "leaf_db")
	leaf := workflow.Define(wiring, "leaf", "TestLeafServiceImplWithDB", leaf_cache, leaf_db)
	nonleaf := workflow.Define(wiring, "nonleaf", "TestNonLeafService", leaf)

	app := assertBuildSuccess(t, wiring, leaf, leaf_db, nonleaf)

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
