package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

const apiKeyPrefix = "sk-"

type SnowflakeGenerator struct {
	mu        sync.Mutex
	lastTime  int64
	sequence  int64
	workerID  int64
}

func NewSnowflakeGenerator(workerID int64) *SnowflakeGenerator {
	return &SnowflakeGenerator{
		workerID: workerID % 1024,
	}
}

func (g *SnowflakeGenerator) Generate() string {
	g.mu.Lock()
	defer g.mu.Unlock()

	now := time.Now().UnixMilli()
	if now == g.lastTime {
		g.sequence++
	} else {
		g.sequence = 0
		g.lastTime = now
	}

	// Combine timestamp + workerID + sequence into a unique value
	unique := (now << 22) | (g.workerID << 12) | g.sequence

	// Convert to bytes and encode as hex
	var buf [8]byte
	for i := 7; i >= 0; i-- {
		buf[i] = byte(unique & 0xFF)
		unique >>= 8
	}

	// Take first 6 bytes for a shorter key
	return apiKeyPrefix + hex.EncodeToString(buf[2:])
}

func GenerateRandomAPIKey() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate random api key: %w", err)
	}
	return apiKeyPrefix + hex.EncodeToString(b), nil
}
