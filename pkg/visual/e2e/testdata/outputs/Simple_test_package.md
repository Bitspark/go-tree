# Package simple

Package simple contains a minimal set of Go declarations for testing.

## Type: SimpleStruct (struct)

SimpleStruct is a basic struct with a few fields.

```go
// SimpleStruct is a basic struct with a few fields.
type SimpleStruct struct {
	Name  string // A string field
	Count int    // An integer field
}
```

### Fields

| Name | Type | Tag | Comment |
|------|------|-----|--------|
| Name | string | `` | A string field |
| Count | int | `` | An integer field |

## Type: SimpleInterface (interface)

SimpleInterface defines a simple interface with one method.

```go
// SimpleInterface defines a simple interface with one method.
type SimpleInterface interface {
	DoSomething() error
}
```

### Methods

| Name | Signature | Comment |
|------|-----------|--------|
| DoSomething | func() error |  |

## Function: SimpleFunction

SimpleFunction is a basic function with no special features.

**Signature:** `func(input string) string`

```go
// SimpleFunction is a basic function with no special features.
func SimpleFunction(input string) string {
	return "Hello, " + input
}
```

## Method: (s *SimpleStruct) SimpleMethod

SimpleMethod is a method defined on SimpleStruct.

**Signature:** `func() string`

```go
// SimpleMethod is a method defined on SimpleStruct.
func (s *SimpleStruct) SimpleMethod() string {
	return s.Name
}
```

