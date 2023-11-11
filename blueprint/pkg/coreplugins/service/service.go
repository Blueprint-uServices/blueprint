package service

// General representation of a service

type (
	ServiceInterface interface {
		GetName() string
		GetMethods() []Method
	}

	Method interface {
		GetName() string
		GetArguments() []Variable
		GetReturns() []Variable
	}

	Variable interface {
		GetName() string
		GetType() string // a "well-known" type
	}
)
