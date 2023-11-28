package tests

import (
	"context"

	"gitlab.mpi-sws.org/cld/blueprint/examples/sockshop/workflow/queuemaster"
	"gitlab.mpi-sws.org/cld/blueprint/runtime/core/registry"
)

// Tests acquire a QueueMaster instance using a service registry.
// This enables us to run local unit tests, while also enabling
// the Blueprint test plugin to auto-generate tests
// for different deployments when compiling an application.
var queuemasterRegistry = registry.NewServiceRegistry[queuemaster.QueueMaster]("queue_master")

func init() {
	// If the tests are run locally, we fall back to this QueueMaster implementation
	queuemasterRegistry.Register("local", func(ctx context.Context) (queuemaster.QueueMaster, error) {
		queue, err := queueRegistry.Get(ctx)
		if err != nil {
			return nil, err
		}

		shipping, err := shippingRegistry.Get(ctx)
		if err != nil {
			return nil, err
		}

		qmaster, err := queuemaster.NewQueueMaster(ctx, queue, shipping)
		if err != nil {
			return nil, err
		}

		// Make sure the queue master is started if it's local
		go func() {
			qmaster.Run(ctx)
		}()

		return qmaster, nil
	})
}
