package service

// General representation of a service

type ServiceInterface interface {
	Name() string
	Methods() []MethodSignature
}

type MethodSignature interface {
	Name() string
	Arguments() []Variable
	Returns() []Variable
}

type Variable interface {
	Name() string
	Type() string // a "well-known" type
}
