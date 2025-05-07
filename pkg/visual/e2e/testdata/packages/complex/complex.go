// Package complex demonstrates a more complex Go package with nested types,
// multiple interfaces, type embedding, and more sophisticated declarations.
package complex

import (
	"context"
	"io"
	"time"
)

// Configuration holds global settings.
type Configuration struct {
	APIKey      string        // API key for external service
	Timeout     time.Duration // Request timeout
	MaxRetries  int           // Maximum number of retries
	Concurrency int           // Number of concurrent operations
	Debug       bool          // Enable debugging output
}

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

// Reader is an interface that reads data from a source.
type Reader interface {
	io.Reader
	io.Closer

	// Reset positions the reader at the beginning.
	Reset() error
}

// Writer is an interface that writes data to a destination.
type Writer interface {
	io.Writer
	io.Closer

	// Flush ensures all data is written.
	Flush() error
}

// DataProcessor processes data from a Reader and writes to a Writer.
type DataProcessor struct {
	reader   Reader
	writer   Writer
	logger   Logger
	config   *Configuration
	handlers []DataHandler
	stats    ProcessingStats
}

// DataHandler is a function type that processes data.
type DataHandler func([]byte) ([]byte, error)

// ProcessingStats tracks statistics about data processing.
type ProcessingStats struct {
	BytesRead    int64
	BytesWritten int64
	StartTime    time.Time
	EndTime      time.Time
	Errors       int
}

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

// NewDataProcessor creates a new DataProcessor with the given configuration.
func NewDataProcessor(reader Reader, writer Writer, logger Logger, config *Configuration) *DataProcessor {
	return &DataProcessor{
		reader: reader,
		writer: writer,
		logger: logger,
		config: config,
	}
}

// AddHandler adds a data handler to the processor chain.
func (p *DataProcessor) AddHandler(handler DataHandler) {
	p.handlers = append(p.handlers, handler)
}

// ProcessDuration returns the total processing duration.
func (p *DataProcessor) ProcessDuration() time.Duration {
	if p.stats.StartTime.IsZero() || p.stats.EndTime.IsZero() {
		return 0
	}
	return p.stats.EndTime.Sub(p.stats.StartTime)
}

// Constants for configuration defaults.
const (
	DefaultTimeout     = 30 * time.Second
	DefaultMaxRetries  = 3
	DefaultConcurrency = 4
	DefaultBufferSize  = 4096
)

// Default configuration values.
var (
	DefaultConfig = &Configuration{
		Timeout:     DefaultTimeout,
		MaxRetries:  DefaultMaxRetries,
		Concurrency: DefaultConcurrency,
		Debug:       false,
	}
)
