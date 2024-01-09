package wiring

import (
	"testing"

	"github.com/blueprint-uservices/blueprint/plugins/simple"
	"github.com/blueprint-uservices/blueprint/plugins/workflow"
)

func TestSimpleNoSQLDB(t *testing.T) {
	spec := newWiringSpec("TestSimpleNoSQLDB")

	leaf_cache := simple.Cache(spec, "leaf_cache")
	leaf_db := simple.NoSQLDB(spec, "leaf_db")
	leaf := workflow.Service(spec, "leaf", "TestLeafServiceImplWithDB", leaf_cache, leaf_db)
	nonleaf := workflow.Service(spec, "nonleaf", "TestNonLeafService", leaf)

	app := assertBuildSuccess(t, spec, leaf, leaf_db, nonleaf)

	assertIR(t, app,
		`TestSimpleNoSQLDB = BlueprintApplication() {
			leaf = TestLeafService(leaf_cache, leaf_db)
			leaf.client = leaf
			leaf.handler.visibility
			leaf_cache = SimpleCache()
			leaf_cache.backend.visibility
			leaf_db = SimpleNoSQLDB()
			leaf_db.backend.visibility
			nonleaf = TestNonLeafService(leaf.client)
			nonleaf.client = nonleaf
			nonleaf.handler.visibility
		  }`)
}
