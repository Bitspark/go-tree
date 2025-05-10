# listfunc

A command-line utility to list all functions in a Go module using the Go-Tree framework.

## Usage

```
go run cmd/listfunc/main.go [module_path]
```

Where:
- `module_path` is the path to a directory containing a Go module (with a go.mod file)
- If `module_path` is not provided, the current directory (`.`) is used

## Examples

### List functions in the current directory
```
go run cmd/listfunc/main.go
```

### List functions in a specific module
```
go run cmd/listfunc/main.go /path/to/your/module
```

## Output Format

The output displays all functions found in the module, grouped by package:

```
Module: github.com/example/module
Directory: /path/to/your/module

Functions:

[github.com/example/module/pkg1]
  Function1(arg1 string, arg2 int) string
  Function2() error

[github.com/example/module/pkg2]
  AnotherFunction(data []byte) (int, error)

Total functions: 3
``` 