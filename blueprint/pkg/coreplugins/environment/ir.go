package environment

import "github.com/blueprint-uservices/blueprint/blueprint/pkg/ir"

type EnvVarConfig struct {
	ir.IRConfig
	Key string
	Val string
}

type EnvVarNode struct {
	Env          EnvVarConfig
	InstanceName string
}

func (e *EnvVarNode) Name() string {
	return e.InstanceName + "." + e.Env.Key
}

func (e *EnvVarNode) String() string {
	return e.InstanceName + "." + e.Env.Key
}

func (e *EnvVarNode) ImplementsIRMetadata() {}

func (e *EnvVarConfig) Name() string {
	return e.Key
}

func (e *EnvVarConfig) String() string {
	return e.Key + " = EnvVarConfig()"
}

func (e *EnvVarConfig) Optional() bool {
	return false
}

func (e *EnvVarConfig) HasValue() bool {
	return e.Val != ""
}

func (e *EnvVarConfig) Value() string {
	return e.Val
}

func (e *EnvVarConfig) ImplementsIRConfig() {}
