package hello

import (
	"context"

	"go-fiber/pkg/db"
)

type Service struct {
	db db.DB
}

func NewService(db db.DB) *Service {
	return &Service{db: db}
}

// PingStatus indica o resultado do ping.
type PingStatus string

const (
	PingOK            PingStatus = "ok"
	PingDBDisabled    PingStatus = "db_disabled"
	PingDBUnavailable PingStatus = "db_unavailable"
)

// Ping verifica conectividade com o banco.
func (s *Service) Ping(ctx context.Context) PingStatus {
	if s.db == nil {
		return PingDBDisabled
	}
	if err := s.db.Ping(ctx); err != nil {
		return PingDBUnavailable
	}
	return PingOK
}
