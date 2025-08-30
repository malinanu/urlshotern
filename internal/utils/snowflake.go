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