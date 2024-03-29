<!-- Code generated by gomarkdoc. DO NOT EDIT -->

# config

```go
import "github.com/blueprint-uservices/blueprint/examples/train_ticket/workflow/config"
```

package config implements ts\-config\-service from the train ticket application

## Index

- [type Config](<#Config>)
- [type ConfigService](<#ConfigService>)
- [type ConfigServiceImpl](<#ConfigServiceImpl>)
  - [func NewConfigServiceImpl\(ctx context.Context, db backend.NoSQLDatabase\) \(\*ConfigServiceImpl, error\)](<#NewConfigServiceImpl>)
  - [func \(c \*ConfigServiceImpl\) Create\(ctx context.Context, conf Config\) error](<#ConfigServiceImpl.Create>)
  - [func \(c \*ConfigServiceImpl\) Delete\(ctx context.Context, name string\) error](<#ConfigServiceImpl.Delete>)
  - [func \(c \*ConfigServiceImpl\) Find\(ctx context.Context, name string\) \(Config, error\)](<#ConfigServiceImpl.Find>)
  - [func \(c \*ConfigServiceImpl\) FindAll\(ctx context.Context\) \(\[\]Config, error\)](<#ConfigServiceImpl.FindAll>)
  - [func \(c \*ConfigServiceImpl\) Update\(ctx context.Context, conf Config\) \(bool, error\)](<#ConfigServiceImpl.Update>)


<a name="Config"></a>
## type [Config](<https://github.com/blueprint-uservices/blueprint/blob/main/examples/train_ticket/workflow/config/data.go#L3-L7>)



```go
type Config struct {
    Name        string
    Value       string
    Description string
}
```

<a name="ConfigService"></a>
## type [ConfigService](<https://github.com/blueprint-uservices/blueprint/blob/main/examples/train_ticket/workflow/config/configService.go#L13-L24>)

Config Service manages Config variables in the application

```go
type ConfigService interface {
    // Creates a new config variable
    Create(ctx context.Context, conf Config) error
    // Updates an existing `conf` config variable
    Update(ctx context.Context, conf Config) (bool, error)
    // Find a config variable using its `name`
    Find(ctx context.Context, name string) (Config, error)
    // Deletes an existing config variable using its `name`
    Delete(ctx context.Context, name string) error
    // Find all config variables
    FindAll(ctx context.Context) ([]Config, error)
}
```

<a name="ConfigServiceImpl"></a>
## type [ConfigServiceImpl](<https://github.com/blueprint-uservices/blueprint/blob/main/examples/train_ticket/workflow/config/configService.go#L27-L29>)

Implementation of Config Service

```go
type ConfigServiceImpl struct {
    // contains filtered or unexported fields
}
```

<a name="NewConfigServiceImpl"></a>
### func [NewConfigServiceImpl](<https://github.com/blueprint-uservices/blueprint/blob/main/examples/train_ticket/workflow/config/configService.go#L32>)

```go
func NewConfigServiceImpl(ctx context.Context, db backend.NoSQLDatabase) (*ConfigServiceImpl, error)
```

Creates a new ConfigService object

<a name="ConfigServiceImpl.Create"></a>
### func \(\*ConfigServiceImpl\) [Create](<https://github.com/blueprint-uservices/blueprint/blob/main/examples/train_ticket/workflow/config/configService.go#L36>)

```go
func (c *ConfigServiceImpl) Create(ctx context.Context, conf Config) error
```



<a name="ConfigServiceImpl.Delete"></a>
### func \(\*ConfigServiceImpl\) [Delete](<https://github.com/blueprint-uservices/blueprint/blob/main/examples/train_ticket/workflow/config/configService.go#L87>)

```go
func (c *ConfigServiceImpl) Delete(ctx context.Context, name string) error
```



<a name="ConfigServiceImpl.Find"></a>
### func \(\*ConfigServiceImpl\) [Find](<https://github.com/blueprint-uservices/blueprint/blob/main/examples/train_ticket/workflow/config/configService.go#L66>)

```go
func (c *ConfigServiceImpl) Find(ctx context.Context, name string) (Config, error)
```



<a name="ConfigServiceImpl.FindAll"></a>
### func \(\*ConfigServiceImpl\) [FindAll](<https://github.com/blueprint-uservices/blueprint/blob/main/examples/train_ticket/workflow/config/configService.go#L96>)

```go
func (c *ConfigServiceImpl) FindAll(ctx context.Context) ([]Config, error)
```



<a name="ConfigServiceImpl.Update"></a>
### func \(\*ConfigServiceImpl\) [Update](<https://github.com/blueprint-uservices/blueprint/blob/main/examples/train_ticket/workflow/config/configService.go#L57>)

```go
func (c *ConfigServiceImpl) Update(ctx context.Context, conf Config) (bool, error)
```



Generated by [gomarkdoc](<https://github.com/princjef/gomarkdoc>)
