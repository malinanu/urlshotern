package utils

import (
	"sync"
	"time"

	"github.com/sony/sonyflake"
)

var (
	sf   *sonyflake.Sonyflake
	once sync.Once
)

// SnowflakeGenerator wraps a snowflake ID generator
type SnowflakeGenerator struct {
	sf *sonyflake.Sonyflake
}

// NewSnowflakeGenerator creates a new SnowflakeGenerator
func NewSnowflakeGenerator(nodeID int64) (*SnowflakeGenerator, error) {
	settings := sonyflake.Settings{
		MachineID: func() (uint16, error) {
			return uint16(nodeID), nil
		},
		StartTime: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	}
	sf := sonyflake.NewSonyflake(settings)
	if sf == nil {
		return nil, ErrSnowflakeInitFailed
	}
	return &SnowflakeGenerator{sf: sf}, nil
}

// Generate generates a unique snowflake ID
func (g *SnowflakeGenerator) Generate() (int64, error) {
	id, err := g.sf.NextID()
	if err != nil {
		return 0, err
	}
	return int64(id), nil
}

// InitializeSnowflake initializes the Snowflake ID generator
func InitializeSnowflake(nodeID int64) error {
	var err error
	once.Do(func() {
		settings := sonyflake.Settings{
			MachineID: func() (uint16, error) {
				return uint16(nodeID), nil
			},
			StartTime: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		}
		sf = sonyflake.NewSonyflake(settings)
		if sf == nil {
			err = ErrSnowflakeInitFailed
		}
	})
	return err
}

// GenerateID generates a unique snowflake ID
func GenerateID() (int64, error) {
	if sf == nil {
		return 0, ErrSnowflakeNotInitialized
	}

	id, err := sf.NextID()
	if err != nil {
		return 0, err
	}

	return int64(id), nil
}

// GenerateSnowflakeID is an alias for GenerateID for compatibility
func GenerateSnowflakeID() (int64, error) {
	return GenerateID()
}

// Custom errors
var (
	ErrSnowflakeInitFailed      = &SnowflakeError{Message: "failed to initialize Snowflake"}
	ErrSnowflakeNotInitialized  = &SnowflakeError{Message: "Snowflake not initialized"}
)

type SnowflakeError struct {
	Message string
}

func (e *SnowflakeError) Error() string {
	return e.Message
}