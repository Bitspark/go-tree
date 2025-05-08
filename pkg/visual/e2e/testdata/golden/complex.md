# Package complex

Package complex demonstrates a more complex Go package with nested types,
multiple interfaces, type embedding, and more sophisticated declarations.

## Type: Configuration (struct)

Configuration holds global settings.

```go
// Configuration holds global settings.
type Configuration struct {
	APIKey      string        // API key for external service
	Timeout     time.Duration // Request timeout
	MaxRetries  int           // Maximum number of retries
	Concurrency int           // Number of concurrent operations
	Debug       bool          // Enable debugging output
}
```

### Fields

| Name | Type | Tag | Comment |
|------|------|-----|--------|
| APIKey | string | `` | API key for external service |
| Timeout | time.Duration | `` | Request timeout |
| MaxRetries | int | `` | Maximum number of retries |
| Concurrency | int | `` | Number of concurrent operations |
| Debug | bool | `` | Enable debugging output |

## Type: Logger (interface)

Logger represents a logging interface.

```go
// Logger represents a logging interface.
type Logger interface {
	// Info logs informational messages.
	Info(msg string, args ...any)

	// Error logs error messages.
	Error(err error, msg string, args ...any)

	// Debug logs debug messages when debugging is enabled.
	Debug(msg string, args ...any)

	// WithField returns a new logger with the field attached.
	WithField(key string, value any) Logger
}
```

### Methods

| Name | Signature | Comment |
|------|-----------|--------|
| Info | func(msg string, args ...any) |  |
| Error | func(err error, msg string, args ...any) |  |
| Debug | func(msg string, args ...any) |  |
| WithField | func(key string, value any) Logger |  |

## Type: Reader (interface)

Reader is an interface that reads data from a source.

```go
// Reader is an interface that reads data from a source.
type Reader interface {
	io.Reader
	io.Closer

	// Reset positions the reader at the beginning.
	Reset() error
}
```

### Methods

| Name | Signature | Comment |
|------|-----------|--------|
| io.Reader | *embedded interface* |  |
| io.Closer | *embedded interface* |  |
| Reset | func() error |  |

## Type: Writer (interface)

Writer is an interface that writes data to a destination.

```go
// Writer is an interface that writes data to a destination.
type Writer interface {
	io.Writer
	io.Closer

	// Flush ensures all data is written.
	Flush() error
}
```

### Methods

| Name | Signature | Comment |
|------|-----------|--------|
| io.Writer | *embedded interface* |  |
| io.Closer | *embedded interface* |  |
| Flush | func() error |  |

## Type: DataProcessor (struct)

DataProcessor processes data from a Reader and writes to a Writer.

```go
// DataProcessor processes data from a Reader and writes to a Writer.
type DataProcessor struct {
	reader   Reader
	writer   Writer
	logger   Logger
	config   *Configuration
	handlers []DataHandler
	stats    ProcessingStats
}
```

### Fields

| Name | Type | Tag | Comment |
|------|------|-----|--------|
| reader | Reader | `` |  |
| writer | Writer | `` |  |
| logger | Logger | `` |  |
| config | *Configuration | `` |  |
| handlers | []DataHandler | `` |  |
| stats | ProcessingStats | `` |  |

## Type: DataHandler (type)

DataHandler is a function type that processes data.

```go
// DataHandler is a function type that processes data.
type DataHandler func([]byte) ([]byte, error)
```

## Type: ProcessingStats (struct)

ProcessingStats tracks statistics about data processing.

```go
// ProcessingStats tracks statistics about data processing.
type ProcessingStats struct {
	BytesRead    int64
	BytesWritten int64
	StartTime    time.Time
	EndTime      time.Time
	Errors       int
}
```

### Fields

| Name | Type | Tag | Comment |
|------|------|-----|--------|
| BytesRead | int64 | `` |  |
| BytesWritten | int64 | `` |  |
| StartTime | time.Time | `` |  |
| EndTime | time.Time | `` |  |
| Errors | int | `` |  |

## Method: (p *DataProcessor) Process

Process reads data from the reader and writes processed data to the writer.

**Signature:** `func(ctx context.Context) (ProcessingStats, error)`

```go
// Process reads data from the reader and writes processed data to the writer.
func (p *DataProcessor) Process(ctx context.Context) (ProcessingStats, error) {
	p.stats.StartTime = time.Now()

	buffer := make([]byte, 4096)
	var err error

	for {
		select {
		case <-ctx.Done():
			p.logger.Info("Processing canceled")
			return p.stats, ctx.Err()
		default:
			// Read data
			n, err := p.reader.Read(buffer)
			if err != nil && err != io.EOF {
				p.stats.Errors++
				p.logger.Error(err, "Error reading data")
				continue
			}

			if n == 0 {
				break
			}

			p.stats.BytesRead += int64(n)
			data := buffer[:n]

			// Apply handlers
			var processErr error
			for _, handler := range p.handlers {
				data, processErr = handler(data)
				if processErr != nil {
					p.stats.Errors++
					p.logger.Error(processErr, "Error processing data")
					break
				}
			}

			if processErr != nil {
				continue
			}

			// Write processed data
			written, err := p.writer.Write(data)
			if err != nil {
				p.stats.Errors++
				p.logger.Error(err, "Error writing data")
				continue
			}

			p.stats.BytesWritten += int64(written)

			if err == io.EOF {
				break
			}
		}
	}

	if flushErr := p.writer.Flush(); flushErr != nil {
		p.logger.Error(flushErr, "Error flushing writer")
		p.stats.Errors++
	}

	p.stats.EndTime = time.Now()
	return p.stats, err
}
```

## Function: NewDataProcessor

NewDataProcessor creates a new DataProcessor with the given configuration.

**Signature:** `func(reader Reader, writer Writer, logger Logger, config *Configuration) *DataProcessor`

```go
// NewDataProcessor creates a new DataProcessor with the given configuration.
func NewDataProcessor(reader Reader, writer Writer, logger Logger, config *Configuration) *DataProcessor {
	return &DataProcessor{
		reader: reader,
		writer: writer,
		logger: logger,
		config: config,
	}
}
```

## Method: (p *DataProcessor) AddHandler

AddHandler adds a data handler to the processor chain.

**Signature:** `func(handler DataHandler)`

```go
// AddHandler adds a data handler to the processor chain.
func (p *DataProcessor) AddHandler(handler DataHandler) {
	p.handlers = append(p.handlers, handler)
}
```

## Method: (p *DataProcessor) ProcessDuration

ProcessDuration returns the total processing duration.

**Signature:** `func() time.Duration`

```go
// ProcessDuration returns the total processing duration.
func (p *DataProcessor) ProcessDuration() time.Duration {
	if p.stats.StartTime.IsZero() || p.stats.EndTime.IsZero() {
		return 0
	}
	return p.stats.EndTime.Sub(p.stats.StartTime)
}
```

