package server

import (
	"github.com/AInicorn/autovibe/database-service/ent"
)

// DatabaseService wraps the Ent client for database operations
type DatabaseService struct {
	client *ent.Client
}

func NewDatabaseService(client *ent.Client) *DatabaseService {
	return &DatabaseService{
		client: client,
	}
}

// Client returns the underlying Ent client
func (s *DatabaseService) Client() *ent.Client {
	return s.client
}

// Close closes the database connection
func (s *DatabaseService) Close() error {
	return s.client.Close()
}