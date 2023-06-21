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

	// Создание таблицы
	err = CreateTable(conn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating table: %v\n", err)
		os.Exit(1)
	}

	err = PrintAllURLs(conn)
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
			ID TEXT PRIMARY KEY,
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

func PrintAllURLs(conn *pgx.Conn) error {
	rows, err := conn.Query(context.Background(), "SELECT * FROM urls")
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var id, url string
		err := rows.Scan(&id, &url)
		if err != nil {
			return err
		}
		fmt.Printf("ID: %s, URL: %s\n", id, url)
	}

	if err := rows.Err(); err != nil {
		return err
	}

	return nil
}
