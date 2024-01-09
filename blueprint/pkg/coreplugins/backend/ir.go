// Package backend provides IR node interfaces for common backend components.
package backend

import (
	"github.com/Blueprint-uServices/blueprint/blueprint/pkg/coreplugins/service"
	"github.com/Blueprint-uServices/blueprint/blueprint/pkg/ir"
)

type (
	Cache interface {
		ir.IRNode
		service.ServiceNode
	}

	NoSQLDB interface {
		ir.IRNode
		service.ServiceNode
	}

	Queue interface {
		ir.IRNode
		service.ServiceNode
	}

	RelDB interface {
		ir.IRNode
		service.ServiceNode
	}
)
