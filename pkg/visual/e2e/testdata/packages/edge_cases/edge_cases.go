// Package edge_cases demonstrates various Go language edge cases and newer features
// that might challenge formatters and visualizers.
package edge_cases

import (
	"context"
	"fmt"
	"reflect"
	"sync"
)

// Generic type with type constraints
type Stack[T any] struct {
	items []T
	mutex sync.Mutex
}

// Method on generic type
func (s *Stack[T]) Push(item T) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.items = append(s.items, item)
}

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

// Generic interface
type Mapper[T, U any] interface {
	Map(input T) U
}

// Implementing the generic interface
type StringToIntMapper struct{}

func (m StringToIntMapper) Map(input string) int {
	return len(input)
}

// Complex embedding with multiple layers
type BaseComponent struct {
	ID   string
	Name string
}

// First level embedding
type InteractiveComponent struct {
	BaseComponent
	Enabled bool
	Handler func() error
}

// Second level embedding with embedding from elsewhere
type ButtonComponent struct {
	InteractiveComponent
	Stack[string] // Embedded generic type
	Label         string
	Size          string
}

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

// Struct with struct tags and complex field types
type ConfigurationObject struct {
	Values       map[string]interface{}          `json:"values" yaml:"values"`
	Handlers     []func(string) error            `json:"-" yaml:"-"`
	Cache        sync.Map                        `json:"-" yaml:"-"`
	GenericField Stack[map[string]interface{}]   `json:"generic_field,omitempty"`
	Callback     func(ctx context.Context) error `json:"-"`
}

// Interface embedding multiple interfaces
type ComplexHandler interface {
	fmt.Stringer
	error
	Handle(ctx context.Context, input interface{}) (interface{}, error)
}

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

// Type alias vs type definition
type MyString string
type AliasString = string

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

// Complex type using generics and reflection
type DynamicRegistry[K comparable, V any] struct {
	values   map[K]V
	mu       sync.RWMutex
	watchers []func(K, V)
	typeInfo reflect.Type
}

func NewDynamicRegistry[K comparable, V any]() *DynamicRegistry[K, V] {
	var v V
	return &DynamicRegistry[K, V]{
		values:   make(map[K]V),
		typeInfo: reflect.TypeOf(v),
	}
}

// Embedded struct with unexported fields and a complex method
type configuration struct {
	settings map[string]interface{}
	once     sync.Once
	mu       sync.RWMutex
}

// Complex struct with embedded unexported struct
type ServiceManager struct {
	configuration

	Services []interface{}
	Contexts []context.Context
}

// Method on struct with embedded unexported fields
func (s *ServiceManager) Configure(settings map[string]interface{}) {
	s.once.Do(func() {
		s.mu.Lock()
		defer s.mu.Unlock()
		s.settings = settings
	})
}
