package db

import (
	"context"
	"fmt"
	"github.com/egosha7/shortlink/internal/config"
	"github.com/jackc/pgx/v4"
	"os"
)

func ConnectToDB(cfg *config.Config) (*pgx.Conn, error) {

	if cfg.DataBase == "" {
		// Возвращаем nil, если строка подключения пуста
		return nil, nil
	}

	connConfig, err := pgx.ParseConfig(cfg.DataBase)
	if err != nil {
		return nil, err
	}

	conn, err := pgx.ConnectConfig(context.Background(), connConfig)
	if err != nil {
		return nil, err
	}

	// Создание таблицы
	err = CreateTable(conn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating table: %v\n", err)
		os.Exit(1)
	}

	return conn, nil
}

func CreateTable(conn *pgx.Conn) error {
	_, err := conn.Exec(
		context.Background(), `
		CREATE TABLE IF NOT EXISTS urls (
			ID SERIAL PRIMARY KEY,
			SHORTURL TEXT,
			URL TEXT,
			UNIQUE (URL)
		)
	`,
	)
	if err != nil {
		return err
	}

	return nil
}
