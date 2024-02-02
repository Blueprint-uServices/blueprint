package address

import (
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/blueprint"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/ir"
)

// Removes and returns all [BindConfig] and [DialConfig] nodes from the provided list of nodes,
// as well as the remaining nodes.
func Split(nodes []ir.IRNode) (binds []*BindConfig, dials []*DialConfig, remaining []ir.IRNode) {
	for _, node := range nodes {
		switch n := node.(type) {
		case *BindConfig:
			binds = append(binds, n)
		case *DialConfig:
			dials = append(dials, n)
		default:
			remaining = append(remaining, node)
		}
	}
	return
}

// Assigns the provided hostname to all [BindConfig] nodes in the provided list of nodes.
func SetHostname(hostname string, nodes []*BindConfig) {
	for _, addr := range nodes {
		addr.Hostname = hostname
	}
}

// Removes the hostname and port assignments from a node
func Clear(binds []*BindConfig) {
	for _, bind := range binds {
		bind.Hostname = ""
		bind.Port = 0
	}
}

// For any of the provided [BindConfig] nodes, if they have not already got a port assigned
// to them, then this method will assign a port.  Ports will be assigned so as not to use
// any port already assigned.  If the same port has been assigned multiple times already,
// an error will be returned.
//
// Ports will be assigned using the PreferredPort field.
//
// Returns preassigned, the list of [BindConfig] nodes that already had a port assigned;
// assigned, the list of [BindConfig] nodes that were assigned a port; and
// err if there was a collision in the preassigned ports
func AssignPorts(binds []*BindConfig) (preassigned []*BindConfig, assigned []*BindConfig, err error) {
	ports := make(map[uint16]*BindConfig)

	// Save any pre-assigned ports
	for _, bind := range binds {
		if bind.Port != 0 {
			if other, conflict := ports[bind.Port]; conflict {
				err = blueprint.Errorf("%v and %v both pre-assigned to port %v", other.Name(), bind.Name(), bind.Port)
				return
			}
			ports[bind.Port] = bind
			preassigned = append(preassigned, bind)
		} else {
			assigned = append(assigned, bind)
		}
	}

	// Assign remaining ports
	for _, addr := range assigned {
		candidatePort := addr.PreferredPort
		if candidatePort == 0 {
			candidatePort = 2000
		}
		for {
			if _, alreadyAssigned := ports[candidatePort]; !alreadyAssigned {
				addr.Port = candidatePort
				ports[addr.Port] = addr
				break
			}
			candidatePort++
		}
	}

	// Set the preferred ports of all addresses
	for _, bind := range binds {
		bind.PreferredPort = bind.Port
	}
	return
}
