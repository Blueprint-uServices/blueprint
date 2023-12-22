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
func TestContainerModifierInstantiation(t *testing.T) {
	spec := newWiringSpec("TestContainerModifierInstantiation")

	leaf := workflow.Service(spec, "leaf", "TestLeafServiceImpl")
	nonleaf := workflow.Service(spec, "nonleaf", "TestNonLeafService", leaf)

	grpc.Deploy(spec, leaf)
	goproc.Deploy(spec, leaf)
	linuxcontainer.Deploy(spec, leaf)

	grpc.Deploy(spec, nonleaf)
	goproc.Deploy(spec, nonleaf)
	linuxcontainer.Deploy(spec, nonleaf)

	app := assertBuildSuccess(t, spec, nonleaf)

	assertIR(t, app,
		`TestContainerModifierInstantiation = BlueprintApplication() {
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
            nonleaf.client = nonleaf.grpc_client
            nonleaf.grpc.addr
            nonleaf.grpc.bind_addr = AddressConfig()
            nonleaf.grpc.dial_addr = AddressConfig()
            nonleaf.grpc_client = GRPCClient(nonleaf.grpc.dial_addr)
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

func TestContainerMixedInstantiation(t *testing.T) {
	spec := newWiringSpec("TestContainerMixedInstantiation")

	leaf := workflow.Service(spec, "leaf", "TestLeafServiceImpl")
	nonleaf := workflow.Service(spec, "nonleaf", "TestNonLeafService", leaf)

	grpc.Deploy(spec, leaf)
	goproc.Deploy(spec, leaf)
	linuxcontainer.Deploy(spec, leaf)

	grpc.Deploy(spec, nonleaf)
	nonleafproc := goproc.CreateProcess(spec, "nonleaf_proc", nonleaf)
	nonleafctr := linuxcontainer.CreateContainer(spec, "nonleaf_ctr", nonleafproc)

	app := assertBuildSuccess(t, spec, nonleafctr)

	assertIR(t, app,
		`TestContainerMixedInstantiation = BlueprintApplication() {
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

func TestContainerExplicitInstantiation(t *testing.T) {
	spec := newWiringSpec("TestContainerExplicitInstantiation")

	leaf := workflow.Service(spec, "leaf", "TestLeafServiceImpl")
	nonleaf := workflow.Service(spec, "nonleaf", "TestNonLeafService", leaf)

	grpc.Deploy(spec, leaf)
	goproc.Deploy(spec, leaf)
	linuxcontainer.Deploy(spec, leaf)

	grpc.Deploy(spec, nonleaf)
	nonleafproc := goproc.CreateProcess(spec, "nonleaf_proc", nonleaf)
	nonleafctr := linuxcontainer.CreateContainer(spec, "nonleaf_ctr", nonleafproc)

	app := assertBuildSuccess(t, spec, leaf+"_ctr", nonleafctr)

	assertIR(t, app,
		`TestContainerExplicitInstantiation = BlueprintApplication() {
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

func TestContainerExplicitNamespaceInstantiation(t *testing.T) {
	spec := newWiringSpec("TestContainerExplicitNamespaceInstantiation")

	leaf := workflow.Service(spec, "leaf", "TestLeafServiceImpl")
	nonleaf := workflow.Service(spec, "nonleaf", "TestNonLeafService", leaf)

	grpc.Deploy(spec, leaf)
	leafproc := goproc.CreateProcess(spec, "leaf_proc", leaf)
	leafctr := linuxcontainer.CreateContainer(spec, "leaf_ctr", leafproc)

	grpc.Deploy(spec, nonleaf)
	nonleafproc := goproc.CreateProcess(spec, "nonleaf_proc", nonleaf)
	nonleafctr := linuxcontainer.CreateContainer(spec, "nonleaf_ctr", nonleafproc)

	app := assertBuildSuccess(t, spec, nonleafctr, leafctr)

	assertIR(t, app,
		`TestContainerExplicitNamespaceInstantiation = BlueprintApplication() {
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
