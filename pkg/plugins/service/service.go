package service

// General representation of a service

type Variable struct {
	Name string
	Type string
}

type ServiceMethodDeclaration struct {
	Name string
	Args []Variable
}

type ServiceInterface struct {
	Name            string
	ConstructorArgs []Variable
	Methods         []ServiceMethodDeclaration
}
