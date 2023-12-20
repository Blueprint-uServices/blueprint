package wiring

import (
	"testing"

	"gitlab.mpi-sws.org/cld/blueprint/plugins/goproc"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/grpc"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/linuxcontainer"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/workflow"
)

func TestContainerModifier(t *testing.T) {
	spec := newWiringSpec("TestContainerModifier")

	leaf := workflow.Service(spec, "leaf", "TestLeafServiceImpl")
	nonleaf := workflow.Service(spec, "nonleaf", "TestNonLeafService", leaf)

	grpc.Deploy(spec, leaf)
	goproc.Deploy(spec, leaf)
	linuxcontainer.Deploy(spec, leaf)

	grpc.Deploy(spec, nonleaf)
	goproc.Deploy(spec, nonleaf)
	linuxcontainer.Deploy(spec, nonleaf)

	app := assertBuildSuccess(t, spec, nonleaf+"_ctr")

	assertIR(t, app,
		`TestContainerModifier = BlueprintApplication() {
			leaf.grpc.addr
			leaf.grpc.bind_addr = AddressConfig()
			leaf.grpc.dial_addr = AddressConfig()
			leaf.handler.visibility
			leaf_ctr = LinuxContainer(leaf.grpc.bind_addr) {
			  leaf_proc = GolangProcessNode(leaf.grpc.bind_addr) {
				leaf = TestLeafService()
				leaf.grpc_server = GRPCServer(leaf, leaf.grpc.bind_addr)
			  }
			}
			nonleaf.grpc.addr
			nonleaf.grpc.bind_addr = AddressConfig()
			nonleaf.handler.visibility
			nonleaf_ctr = LinuxContainer(leaf.grpc.dial_addr, nonleaf.grpc.bind_addr) {
			  nonleaf_proc = GolangProcessNode(leaf.grpc.dial_addr, nonleaf.grpc.bind_addr) {
				leaf.client = leaf.grpc_client
				leaf.grpc_client = GRPCClient(leaf.grpc.dial_addr)
				nonleaf = TestNonLeafService(leaf.client)
				nonleaf.grpc_server = GRPCServer(nonleaf, nonleaf.grpc.bind_addr)
			  }
			}
		  }`)
}
