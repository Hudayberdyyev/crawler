package postgres

import (
	"context"
	"github.com/jackc/pgx"
)

type Config struct {
	Host string
	Port uint16
	Username string
	Password string
	DBName string
	SSLMode string
}

func NewPostgresDB(cfg Config) (*pgx.Conn, error) {
	db, err := pgx.Connect(pgx.ConnConfig{
		Host:                 cfg.Host,
		Port:				  cfg.Port,
		Database:             cfg.DBName,
		User:                 cfg.Username,
		Password:             cfg.Password,
	})

	if err != nil {
		return nil, err
	}

	err = db.Ping(context.Background())
	if err != nil {
		return nil, err
	}

	return db, nil
}