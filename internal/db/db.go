package db

import (
	"context"
	"github.com/egosha7/shortlink/internal/config"
	"github.com/jackc/pgx/v4"
)

func ConnectToDB(cfg *config.Config) (*pgx.Conn, error) {

	if cfg.DataBase == "" {
		// Возвращаем nil, если строка подключения пуста
		conn := &pgx.Conn{}
		return conn, nil
	}

	connConfig, err := pgx.ParseConfig(cfg.DataBase)
	if err != nil {
		return nil, err
	}

	conn, err := pgx.ConnectConfig(context.Background(), connConfig)
	if err != nil {
		return nil, err
	}

	return conn, nil
}
