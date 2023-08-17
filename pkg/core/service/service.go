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

func (method ServiceMethodDeclaration) AddArg(Name, Type string) ServiceMethodDeclaration {
	v := Variable{Name, Type}
	method.Args = append(method.Args, v)
	return method
}

type ServiceInterface struct {
	Name            string
	ConstructorArgs []Variable
	Methods         []ServiceMethodDeclaration
}

func (iface ServiceInterface) ExtendMethodArgs(Name, Type string) ServiceInterface {
	// TODO: pprobably this is incorrect and modifies args in-place; check
	v := Variable{}
	v.Name = Name
	v.Type = Type
	for _, method := range iface.Methods {
		method.Args = append(method.Args, v)
	}
	return iface
}
