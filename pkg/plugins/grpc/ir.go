package grpc

// type GolangGRPCServerNode struct {
// 	golang.Node
// 	golang.ArtifactGenerator
// 	golang.CodeGenerator

// 	InstanceName string
// 	Wrapped      golang.Service
// }

// type GolangGRPCClientNode struct {
// 	golang.Node
// 	golang.ArtifactGenerator
// 	golang.CodeGenerator
// 	golang.Service

// 	InstanceName   string
// 	ServiceDetails golang.GolangServiceDetails
// 	ServerPointer  blueprint.IRPointer
// }

// func newGolangGRPCServerNode(name string, wrapped golang.Service) *GolangGRPCServerNode {
// 	node := GolangGRPCServerNode{}
// 	node.InstanceName = name
// 	node.Wrapped = wrapped
// 	return &node
// }

// func newGolangGRPCClientNode(name string, serverPointer blueprint.IRPointer) *GolangGRPCClientNode {
// 	node := GolangGRPCClientNode{}
// 	node.InstanceName = name
// 	node.ServerPointer = serverPointer

// 	// TODO package and files correctly
// 	node.ServiceDetails.Package = "TODO"
// 	node.ServiceDetails.Files = []string{}
// 	node.ServiceDetails.Interface.Name = name
// 	constructorArg := service.Variable{}
// 	constructorArg.Name = "RemoteAddr"
// 	constructorArg.Type = "string"
// 	node.ServiceDetails.Interface.ConstructorArgs = []service.Variable{constructorArg}

// 	return &node
// }

// func (client *GolangGRPCClientNode) SetInterface(node golang.Service) {
// 	client.ServiceDetails.Interface.Methods = node.GetInterface().Methods
// }

// func (n *GolangGRPCServerNode) String() string {
// 	return n.InstanceName
// }

// func (n *GolangGRPCServerNode) Name() string {
// 	return n.InstanceName
// }

// func (n *GolangGRPCClientNode) String() string {
// 	return n.InstanceName
// }

// func (n *GolangGRPCClientNode) Name() string {
// 	return n.InstanceName
// }
