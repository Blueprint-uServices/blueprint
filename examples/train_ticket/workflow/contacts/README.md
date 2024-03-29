<!-- Code generated by gomarkdoc. DO NOT EDIT -->

# contacts

```go
import "github.com/blueprint-uservices/blueprint/examples/train_ticket/workflow/contacts"
```

## Index

- [Constants](<#constants>)
- [type Contact](<#Contact>)
- [type ContactsService](<#ContactsService>)
- [type ContactsServiceImpl](<#ContactsServiceImpl>)
  - [func NewContactsServiceImpl\(ctx context.Context, db backend.NoSQLDatabase\) \(\*ContactsServiceImpl, error\)](<#NewContactsServiceImpl>)
  - [func \(c \*ContactsServiceImpl\) CreateContacts\(ctx context.Context, contact Contact\) error](<#ContactsServiceImpl.CreateContacts>)
  - [func \(c \*ContactsServiceImpl\) Delete\(ctx context.Context, contact Contact\) error](<#ContactsServiceImpl.Delete>)
  - [func \(c \*ContactsServiceImpl\) FindContactsByAccountId\(ctx context.Context, id string\) \(\[\]Contact, error\)](<#ContactsServiceImpl.FindContactsByAccountId>)
  - [func \(c \*ContactsServiceImpl\) FindContactsById\(ctx context.Context, id string\) \(Contact, error\)](<#ContactsServiceImpl.FindContactsById>)
  - [func \(c \*ContactsServiceImpl\) GetAllContacts\(ctx context.Context\) \(\[\]Contact, error\)](<#ContactsServiceImpl.GetAllContacts>)
  - [func \(c \*ContactsServiceImpl\) Modify\(ctx context.Context, contact Contact\) \(bool, error\)](<#ContactsServiceImpl.Modify>)


## Constants

<a name="NULL"></a>DocumentType enum

```go
const (
    NULL int64 = iota
    ID_CARD
    PASSPORT
    OTHER
)
```

<a name="Contact"></a>
## type [Contact](<https://github.com/blueprint-uservices/blueprint/blob/main/examples/train_ticket/workflow/contacts/data.go#L11-L18>)



```go
type Contact struct {
    ID             string
    AccountID      string
    Name           string
    DocumentType   int
    DocumentNumber string
    PhoneNumber    string
}
```

<a name="ContactsService"></a>
## type [ContactsService](<https://github.com/blueprint-uservices/blueprint/blob/main/examples/train_ticket/workflow/contacts/contactsService.go#L12-L25>)

Contacts Service manages contacts for users

```go
type ContactsService interface {
    // Find a contact using its `id`
    FindContactsById(ctx context.Context, id string) (Contact, error)
    // Find all contacts associated with an account with ID `id`
    FindContactsByAccountId(ctx context.Context, id string) ([]Contact, error)
    // Create a new contact
    CreateContacts(ctx context.Context, c Contact) error
    // Delete an existing contact
    Delete(ctx context.Context, c Contact) error
    // Get all existing contacts
    GetAllContacts(ctx context.Context) ([]Contact, error)
    // Modify an existing contact
    Modify(ctx context.Context, contact Contact) (bool, error)
}
```

<a name="ContactsServiceImpl"></a>
## type [ContactsServiceImpl](<https://github.com/blueprint-uservices/blueprint/blob/main/examples/train_ticket/workflow/contacts/contactsService.go#L27-L29>)



```go
type ContactsServiceImpl struct {
    // contains filtered or unexported fields
}
```

<a name="NewContactsServiceImpl"></a>
### func [NewContactsServiceImpl](<https://github.com/blueprint-uservices/blueprint/blob/main/examples/train_ticket/workflow/contacts/contactsService.go#L31>)

```go
func NewContactsServiceImpl(ctx context.Context, db backend.NoSQLDatabase) (*ContactsServiceImpl, error)
```



<a name="ContactsServiceImpl.CreateContacts"></a>
### func \(\*ContactsServiceImpl\) [CreateContacts](<https://github.com/blueprint-uservices/blueprint/blob/main/examples/train_ticket/workflow/contacts/contactsService.go#L74>)

```go
func (c *ContactsServiceImpl) CreateContacts(ctx context.Context, contact Contact) error
```



<a name="ContactsServiceImpl.Delete"></a>
### func \(\*ContactsServiceImpl\) [Delete](<https://github.com/blueprint-uservices/blueprint/blob/main/examples/train_ticket/workflow/contacts/contactsService.go#L95>)

```go
func (c *ContactsServiceImpl) Delete(ctx context.Context, contact Contact) error
```



<a name="ContactsServiceImpl.FindContactsByAccountId"></a>
### func \(\*ContactsServiceImpl\) [FindContactsByAccountId](<https://github.com/blueprint-uservices/blueprint/blob/main/examples/train_ticket/workflow/contacts/contactsService.go#L56>)

```go
func (c *ContactsServiceImpl) FindContactsByAccountId(ctx context.Context, id string) ([]Contact, error)
```



<a name="ContactsServiceImpl.FindContactsById"></a>
### func \(\*ContactsServiceImpl\) [FindContactsById](<https://github.com/blueprint-uservices/blueprint/blob/main/examples/train_ticket/workflow/contacts/contactsService.go#L35>)

```go
func (c *ContactsServiceImpl) FindContactsById(ctx context.Context, id string) (Contact, error)
```



<a name="ContactsServiceImpl.GetAllContacts"></a>
### func \(\*ContactsServiceImpl\) [GetAllContacts](<https://github.com/blueprint-uservices/blueprint/blob/main/examples/train_ticket/workflow/contacts/contactsService.go#L104>)

```go
func (c *ContactsServiceImpl) GetAllContacts(ctx context.Context) ([]Contact, error)
```



<a name="ContactsServiceImpl.Modify"></a>
### func \(\*ContactsServiceImpl\) [Modify](<https://github.com/blueprint-uservices/blueprint/blob/main/examples/train_ticket/workflow/contacts/contactsService.go#L121>)

```go
func (c *ContactsServiceImpl) Modify(ctx context.Context, contact Contact) (bool, error)
```



Generated by [gomarkdoc](<https://github.com/princjef/gomarkdoc>)
