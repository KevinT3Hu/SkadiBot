package utils

import (
	"context"
	"errors"
	"os"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pgvector/pgvector-go"
	pgxvector "github.com/pgvector/pgvector-go/pgx"
)

type DB struct {
	*pgxpool.Pool
}

func NewDB() (*DB, error) {
	username := "umsg"
	passwordFile := os.Getenv("MSG_PASS_FILE")
	if passwordFile == "" {
		return nil, errors.New("MSG_PASS_FILE is empty")
	}
	passwordBytes, err := os.ReadFile(passwordFile)
	if err != nil {
		return nil, err
	}
	password := strings.TrimSpace(string(passwordBytes))
	host := os.Getenv("msg_host")
	port := os.Getenv("msg_port")
	dbname := "msg"
	connStr := "postgres://" + username + ":" + password + "@" + host + ":" + port + "/" + dbname
	ctx := context.Background()
	config, err := pgxpool.ParseConfig(strings.TrimSpace(connStr))
	if err != nil {
		return nil, err
	}
	config.AfterConnect = func(ctx context.Context, c *pgx.Conn) error {
		return pgxvector.RegisterTypes(ctx, c)
	}
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, err
	}
	return &DB{pool}, nil
}

func (db *DB) Close() {
	db.Pool.Close()
}

func (db *DB) SaveMessage(message string, vector []float32, nextMessage string) error {
	if len(vector) != 50 {
		return errors.New("vector length must be 50")
	}
	_, err := db.Exec(context.Background(), "INSERT INTO messages (prev, prev_vector, next) VALUES ($1, $2, $3)", message, pgvector.NewVector(vector), nextMessage)
	return err
}

func (db *DB) MessageExists(message string) (bool, string, error) {
	var exists bool
	var nextMessage string

	err := db.QueryRow(context.Background(), "SELECT EXISTS(SELECT 1 FROM messages WHERE prev = $1)", message).Scan(&exists)
	if err != nil {
		return false, "", err
	}
	if exists {
		// return highest frequency next message
		err = db.QueryRow(context.Background(), "SELECT next FROM messages WHERE prev = $1 GROUP BY next ORDER BY COUNT(*) DESC LIMIT 1", message).Scan(&nextMessage)
	}
	return exists, nextMessage, err
}

func (db *DB) GetNearestMessage(vector []float32) (string, error) {
	if len(vector) != 50 {
		return "", errors.New("vector length must be 50")
	}
	var message string
	err := db.QueryRow(context.Background(), "SELECT next FROM messages ORDER BY prev_vector <-> $1 LIMIT 1", pgvector.NewVector(vector)).Scan(&message)
	return message, err
}

func (db *DB) RebuildMessageVec(rebuildFunc func(string) ([]float32, error)) error {
	rows, err := db.Query(context.Background(), "SELECT prev FROM messages GROUP BY prev")
	if err != nil {
		return err
	}
	defer rows.Close()
	tx, err := db.Pool.BeginTx(context.Background(), pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(context.Background())
	for rows.Next() {
		var message string
		err = rows.Scan(&message)
		if err != nil {
			return err
		}
		vector, err := rebuildFunc(message)
		if err != nil {
			return err
		}
		_, err = tx.Exec(context.Background(), "UPDATE messages SET prev_vector = $1 WHERE prev = $2", pgvector.NewVector(vector), message)
		if err != nil {
			return err
		}
	}
	err = tx.Commit(context.Background())
	return err
}

// db_size, table_size
func (db *DB) GetSize() (string, string, error) {
	var dbSize string
	var tableSize string
	err := db.QueryRow(context.Background(), "SELECT pg_size_pretty(pg_database_size(current_database())), pg_size_pretty(pg_total_relation_size('messages'))").Scan(&dbSize, &tableSize)
	return dbSize, tableSize, err
}

func (db *DB) GetRowCount() (int, error) {
	var count int
	err := db.QueryRow(context.Background(), "SELECT COUNT(*) FROM messages").Scan(&count)
	return count, err
}
