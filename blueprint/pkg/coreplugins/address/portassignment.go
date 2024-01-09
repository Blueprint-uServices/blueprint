package address

import (
	"fmt"
	"strings"

	"github.com/Blueprint-uServices/blueprint/blueprint/pkg/blueprint"
	"github.com/Blueprint-uServices/blueprint/blueprint/pkg/ir"
)

/*
AssignPorts is a helper method intended for use by namespace nodes when they
are compiling code and concrete ports must be assigned to [BindConfig] IR nodes.

The provided nodes can be any IR nodes; this method will filter out only the [BindConfig]
nodes.

Some of the provided nodes might already be assigned to a particular port.  This method
will not change those port assignments, though it will return an error if two nodes
are already pre-assigned to the same port.

Ports will be assigned either ascending from port 2000, or ascending from a node's
preferred port if a preference was specified.

After calling this method, any provided [BindConfig] IR nodes will have their hostname
and port set.
*/
func AssignPorts(hostname string, nodes []ir.IRNode) error {
	// Extract the BindConfig nodes
	addrs := ir.Filter[*BindConfig](nodes)

	ports := make(map[uint16]*BindConfig)

	// Save any pre-assigned ports
	for _, addr := range addrs {
		if addr.Port != 0 {
			if other, conflict := ports[addr.Port]; conflict {
				return blueprint.Errorf("%v and %v both pre-assigned to port %v", other.Name(), addr.Name(), addr.Port)
			}
			ports[addr.Port] = addr
		}
	}

	// Assign preferred ports first
	for _, addr := range addrs {
		if addr.Port == 0 && addr.PreferredPort != 0 {
			if _, conflict := ports[addr.PreferredPort]; !conflict {
				addr.Port = addr.PreferredPort
				addr.Hostname = hostname
				ports[addr.Port] = addr
			}
		}
	}

	// Assign remaining ports
	for _, addr := range addrs {
		if addr.Port == 0 {
			candidatePort := addr.PreferredPort
			if candidatePort == 0 {
				candidatePort = 2000
			}
			for {
				if _, alreadyAssigned := ports[candidatePort]; !alreadyAssigned {
					addr.Port = candidatePort
					addr.Hostname = hostname
					ports[addr.Port] = addr
					break
				}
				candidatePort++
			}

		}
	}

	// Update preferred ports
	for _, addr := range addrs {
		addr.PreferredPort = addr.Port
	}
	return nil
}

// Returns an error if there are [BindConfig] nodes in the provided list that haven't been allocated a port.
func CheckPorts(nodes []ir.IRNode) error {
	var missing []string
	for _, addr := range ir.Filter[*BindConfig](nodes) {
		if addr.Port == 0 {
			missing = append(missing, addr.Name())
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("unassigned bind addresses %v", strings.Join(missing, ", "))
	}
	return nil
}

// Clears the hostname and port from any [BindConfig] node.
//
// This is used by namespace nodes when performing address translation, e.g. between
// ports within a container vs. external to a container.
func ResetPorts(nodes []ir.IRNode) {
	for _, addr := range ir.Filter[*BindConfig](nodes) {
		addr.Port = 0
		addr.Hostname = ""
	}
}
