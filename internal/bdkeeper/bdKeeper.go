package bdkeeper

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/wurt83ow/tinyurl/cmd/shortener/storage"
	"github.com/wurt83ow/tinyurl/internal/models"
)

type Log interface {
	Info(string, ...zapcore.Field)
}
type BDKeeper struct {
	pool *pgxpool.Pool
	log  Log
}

func NewBDKeeper(dns func() string, log Log) *BDKeeper {

	addr := dns()
	if addr == "" {
		log.Info("database dns is empty")
		return nil
	}

	pool, err := pgxpool.New(context.Background(), dns())
	if err != nil {
		log.Info("Unable to connection to database: %v", zap.Error(err))
	}

	err = Bootstrap(pool, log)
	if err != nil {
		log.Info("failed to create database table: %v", zap.Error(err))
		return nil
	}

	log.Info("Connected!")
	return &BDKeeper{
		pool: pool,
		log:  log,
	}
}

func (bdk *BDKeeper) Load() (storage.StorageURL, error) {

	data := make(storage.StorageURL)
	conn, err := bdk.pool.Acquire(context.Background())
	if err != nil {
		bdk.log.Info("Unable to acquire a database connection: %v\n", zap.Error(err))

		return data, err
	}
	defer conn.Release()

	rows, err := conn.Query(context.Background(),
		"SELECT * FROM dataurl")

	if err != nil {
		bdk.log.Info("error while getting data from bd: %v\n", zap.Error(err))

		return data, err
	}

	for rows.Next() {
		record := models.DataURL{}

		s := reflect.ValueOf(&record).Elem()
		numCols := s.NumField()
		columns := make([]interface{}, numCols)
		for i := 0; i < numCols; i++ {
			field := s.Field(i)
			columns[i] = field.Addr().Interface()
		}

		err := rows.Scan(columns...)
		if err != nil {
			log.Fatal(err)
		}
		data[record.ShortURL] = record
	}

	return data, nil
}

func (bdk *BDKeeper) Save(data storage.StorageURL) error {

	conn, err := bdk.pool.Acquire(context.Background())
	if err != nil {
		bdk.log.Info("Unable to acquire a database connection: %v\n", zap.Error(err))

		return err
	}

	defer conn.Release()

	_, err = conn.Exec(context.Background(), "DELETE FROM dataurl")
	if err != nil {
		bdk.log.Info("Unable to DELETE: %v", zap.Error(err))
		return err
	}

	for k, v := range data {
		var id string
		if v.UUID == "" {
			neuuid := uuid.New()
			id = neuuid.String()
		} else {
			id = v.UUID
		}

		row := conn.QueryRow(context.Background(),
			"INSERT INTO dataurl (correlation_id, short_url, original_url) VALUES ($1, $2, $3) RETURNING correlation_id",
			id, k, v.OriginalURL)
		var rowid string
		err = row.Scan(&rowid)
		if err != nil {
			bdk.log.Info("Unable to INSERT: %v", zap.Error(err))
			return err
		}
	}

	return nil
}

func (bdk *BDKeeper) Ping() bool {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	if err := bdk.pool.Ping(ctx); err != nil {
		fmt.Println(err)
		return false
	}

	return true
}

func (bdk *BDKeeper) Close() bool {
	bdk.pool.Close()
	return true
}

func Bootstrap(pool *pgxpool.Pool, log Log) error {

	const query = `
	CREATE TABLE IF NOT EXISTS dataURL (
	correlation_id SERIAL PRIMARY KEY,
	short_url TEXT,
	original_url TEXT
	)`

	conn, err := pool.Acquire(context.Background())
	if err != nil {
		log.Info("Unable to acquire a database connection: %v\n", zap.Error(err))

		return err
	}
	defer conn.Release()

	_, err = conn.Exec(context.Background(), query)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}
