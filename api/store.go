package api

import (
	"maps"
	"slices"
	"strconv"
	"sync"
	"time"

	"github.com/samber/lo"
)

const (
	// max store 100k not to flood in memory store
	maxStoreLimits = 100 * 1000
)

var (
	nextId = 0
)

// Person represents a simple user object
type Person struct {
	ID        string `json:"id,omitempty"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
	UpdatedAt string `json:"updated_at,omitempty"`
}

// In-memory store (map) with mutex
type store struct {
	mu      sync.RWMutex
	records map[string]Person
}

func newStore() *store {
	return &store{
		records: make(map[string]Person),
	}
}

func (s *store) list(start, count int) []Person {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Now it will return random values and that's okay for now
	presons := maps.Values(s.records)
	out := lo.Slice(slices.Collect(presons), start, start+count)

	return out
}

func (s *store) get(id string) (Person, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	p, ok := s.records[id]
	return p, ok
}

func (s *store) create(p Person) Person {
	s.mu.Lock()
	defer s.mu.Unlock()
	//id := uuid.New().String()
	now := time.Now().UTC().Format(time.RFC3339)

	id := strconv.Itoa(nextId)
	nextId = nextId + 1

	p.ID = id
	p.CreatedAt = now
	p.UpdatedAt = now
	s.records[id] = p

	return p
}

func (s *store) update(id string, p Person) (Person, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	existing, ok := s.records[id]
	if !ok {
		return Person{}, false
	}
	// update fields (simple replace of first/last/email)
	if p.FirstName != "" {
		existing.FirstName = p.FirstName
	}
	if p.LastName != "" {
		existing.LastName = p.LastName
	}
	if p.Email != "" {
		existing.Email = p.Email
	}
	existing.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	s.records[id] = existing
	return existing, true
}

func (s *store) delete(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.records[id]; !ok {
		return false
	}
	delete(s.records, id)
	return true
}

func (s *store) count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.records)
}
