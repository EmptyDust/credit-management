package idgen

import (
	"crypto/rand"
	"io"
	"sync"
	"time"

	"github.com/oklog/ulid/v2"
)

type Generator struct {
	mu      sync.Mutex
	entropy io.Reader
}

var defaultGenerator = NewGenerator()

func NewGenerator() *Generator {
	return &Generator{
		entropy: ulid.Monotonic(rand.Reader, 0),
	}
}

func New() string {
	return defaultGenerator.NewAt(time.Now().UTC())
}

func (g *Generator) NewAt(now time.Time) string {
	g.mu.Lock()
	defer g.mu.Unlock()

	if now.IsZero() {
		now = time.Now().UTC()
	}

	return ulid.MustNew(ulid.Timestamp(now), g.entropy).String()
}
