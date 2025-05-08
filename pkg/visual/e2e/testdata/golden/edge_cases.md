# Package edge_cases

Package edge_cases demonstrates various Go language edge cases and newer features
that might challenge formatters and visualizers.

## Type: Stack (struct)

Generic type with type constraints

```go
// Generic type with type constraints
type Stack[T any] struct {
	items []T
	mutex sync.Mutex
}
```

### Fields

| Name | Type | Tag | Comment |
|------|------|-----|--------|
| items | []T | `` |  |
| mutex | sync.Mutex | `` |  |

## Type: Mapper (interface)

Generic interface

```go
// Generic interface
type Mapper[T, U any] interface {
	Map(input T) U
}
```

### Methods

| Name | Signature | Comment |
|------|-----------|--------|
| Map | func(input T) U |  |

## Type: StringToIntMapper (struct)

Implementing the generic interface

```go
// Implementing the generic interface
type StringToIntMapper struct{}
```

## Type: BaseComponent (struct)

Complex embedding with multiple layers

```go
// Complex embedding with multiple layers
type BaseComponent struct {
	ID   string
	Name string
}
```

### Fields

| Name | Type | Tag | Comment |
|------|------|-----|--------|
| ID | string | `` |  |
| Name | string | `` |  |

## Type: InteractiveComponent (struct)

First level embedding

```go
// First level embedding
type InteractiveComponent struct {
	BaseComponent
	Enabled bool
	Handler func() error
}
```

### Fields

| Name | Type | Tag | Comment |
|------|------|-----|--------|
| *embedded* | BaseComponent | `` |  |
| Enabled | bool | `` |  |
| Handler | func() error | `` |  |

## Type: ButtonComponent (struct)

Second level embedding with embedding from elsewhere

```go
// Second level embedding with embedding from elsewhere
type ButtonComponent struct {
	InteractiveComponent
	Stack[string] // Embedded generic type
	Label         string
	Size          string
}
```

### Fields

| Name | Type | Tag | Comment |
|------|------|-----|--------|
| *embedded* | InteractiveComponent | `` |  |
| *embedded* | Stack[string] | `` | Embedded generic type |
| Label | string | `` |  |
| Size | string | `` |  |

## Type: ConfigurationObject (struct)

Struct with struct tags and complex field types

```go
// Struct with struct tags and complex field types
type ConfigurationObject struct {
	Values       map[string]interface{}          `json:"values" yaml:"values"`
	Handlers     []func(string) error            `json:"-" yaml:"-"`
	Cache        sync.Map                        `json:"-" yaml:"-"`
	GenericField Stack[map[string]interface{}]   `json:"generic_field,omitempty"`
	Callback     func(ctx context.Context) error `json:"-"`
}
```

### Fields

| Name | Type | Tag | Comment |
|------|------|-----|--------|
| Values | map[string]interface{} | ``json:"values" yaml:"values"`` |  |
| Handlers | []func(string) error | ``json:"-" yaml:"-"`` |  |
| Cache | sync.Map | ``json:"-" yaml:"-"`` |  |
| GenericField | Stack[map[string]interface{}] | ``json:"generic_field,omitempty"`` |  |
| Callback | func(ctx context.Context) error | ``json:"-"`` |  |

## Type: ComplexHandler (interface)

Interface embedding multiple interfaces

```go
// Interface embedding multiple interfaces
type ComplexHandler interface {
	fmt.Stringer
	error
	Handle(ctx context.Context, input interface{}) (interface{}, error)
}
```

### Methods

| Name | Signature | Comment |
|------|-----------|--------|
| fmt.Stringer | *embedded interface* |  |
| error | *embedded interface* |  |
| Handle | func(ctx context.Context, input interface{}) (interface{}, error) |  |

## Type: MyString (type)

Type alias vs type definition

```go
// Type alias vs type definition
type MyString string
```

## Type: AliasString (alias)

```go
type AliasString = string
```

## Type: DynamicRegistry (struct)

Complex type using generics and reflection

```go
// Complex type using generics and reflection
type DynamicRegistry[K comparable, V any] struct {
	values   map[K]V
	mu       sync.RWMutex
	watchers []func(K, V)
	typeInfo reflect.Type
}
```

### Fields

| Name | Type | Tag | Comment |
|------|------|-----|--------|
| values | map[K]V | `` |  |
| mu | sync.RWMutex | `` |  |
| watchers | []func(K, V) | `` |  |
| typeInfo | reflect.Type | `` |  |

## Type: configuration (struct)

Embedded struct with unexported fields and a complex method

```go
// Embedded struct with unexported fields and a complex method
type configuration struct {
	settings map[string]interface{}
	once     sync.Once
	mu       sync.RWMutex
}
```

### Fields

| Name | Type | Tag | Comment |
|------|------|-----|--------|
| settings | map[string]interface{} | `` |  |
| once | sync.Once | `` |  |
| mu | sync.RWMutex | `` |  |

## Type: ServiceManager (struct)

Complex struct with embedded unexported struct

```go
// Complex struct with embedded unexported struct
type ServiceManager struct {
	configuration

	Services []interface{}
	Contexts []context.Context
}
```

### Fields

| Name | Type | Tag | Comment |
|------|------|-----|--------|
| *embedded* | configuration | `` |  |
| Services | []interface{} | `` |  |
| Contexts | []context.Context | `` |  |

## Method: (s *Stack[T]) Push

Method on generic type

**Signature:** `func(item T)`

```go
// Method on generic type
func (s *Stack[T]) Push(item T) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.items = append(s.items, item)
}
```

## Method: (s *Stack[T]) Pop

Another method on generic type with return value

**Signature:** `func() (T, bool)`

```go
// Another method on generic type with return value
func (s *Stack[T]) Pop() (T, bool) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	var zero T
	if len(s.items) == 0 {
		return zero, false
	}

	lastIndex := len(s.items) - 1
	item := s.items[lastIndex]
	s.items = s.items[:lastIndex]
	return item, true
}
```

## Method: (m StringToIntMapper) Map

**Signature:** `func(input string) int`

```go
func (m StringToIntMapper) Map(input string) int {
	return len(input)
}
```

## Function: ProcessInputs

Function with variadic parameters and complex return type

**Signature:** `func[T comparable](ctx context.Context, defaultValue T, inputs ...T) (results map[T]struct{}, err error)`

```go
// Function with variadic parameters and complex return type
func ProcessInputs[T comparable](ctx context.Context, defaultValue T, inputs ...T) (results map[T]struct{}, err error) {
	results = make(map[T]struct{})

	for _, input := range inputs {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			results[input] = struct{}{}
		}
	}

	if len(results) == 0 {
		results[defaultValue] = struct{}{}
	}

	return results, nil
}
```

## Function: CreateValidator

Function with named return values and closures

**Signature:** `func[T any](rules ...func(T) bool) func(T) (valid bool, failedRules []int)`

```go
// Function with named return values and closures
func CreateValidator[T any](rules ...func(T) bool) func(T) (valid bool, failedRules []int) {
	return func(value T) (valid bool, failedRules []int) {
		valid = true

		for i, rule := range rules {
			if !rule(value) {
				valid = false
				failedRules = append(failedRules, i)
			}
		}

		return
	}
}
```

## Function: DemonstrateAliasVsDefinition

Using alias

**Signature:** `func()`

```go
// Using alias
func DemonstrateAliasVsDefinition() {
	var s1 MyString = "hello"
	var s2 AliasString = "world"

	// This would fail because MyString is a new type
	// s1 = "direct assignment"

	// This works because AliasString is just an alias
	s2 = "direct assignment"

	// Type conversions
	fmt.Println(string(s1), s2)
}
```

## Function: NewDynamicRegistry

**Signature:** `func[K comparable, V any]() *DynamicRegistry[K, V]`

```go
func NewDynamicRegistry[K comparable, V any]() *DynamicRegistry[K, V] {
	var v V
	return &DynamicRegistry[K, V]{
		values:   make(map[K]V),
		typeInfo: reflect.TypeOf(v),
	}
}
```

## Method: (s *ServiceManager) Configure

Method on struct with embedded unexported fields

**Signature:** `func(settings map[string]interface{})`

```go
// Method on struct with embedded unexported fields
func (s *ServiceManager) Configure(settings map[string]interface{}) {
	s.once.Do(func() {
		s.mu.Lock()
		defer s.mu.Unlock()
		s.settings = settings
	})
}
```

