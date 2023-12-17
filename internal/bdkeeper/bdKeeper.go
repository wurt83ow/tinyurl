// Package bdkeeper provides a PostgreSQL-backed implementation of the storage.Keeper interface.
// It manages the storage and retrieval of URL and user data in a PostgreSQL database.
package bdkeeper

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/wurt83ow/tinyurl/internal/models"
	"github.com/wurt83ow/tinyurl/internal/storage"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Log is an interface representing a logger with Info method.
type Log interface {
	Info(string, ...zapcore.Field)
}

// BDKeeper is a PostgreSQL-backed implementation of the storage.Keeper interface.
type BDKeeper struct {
	conn *sql.DB
	log  Log
}

// NewBDKeeper creates a new BDKeeper instance with the provided DSN (data source name) function and logger.
// It establishes a connection to the PostgreSQL database, performs any required migrations, and returns the BDKeeper instance.
func NewBDKeeper(dsn func() string, log Log) *BDKeeper {
	addr := dsn()
	if addr == "" {
		log.Info("database dsn is empty")
		return nil
	}

	conn, err := sql.Open("pgx", dsn())
	if err != nil {
		log.Info("Unable to connect to the database: ", zap.Error(err))
		return nil
	}

	driver, err := postgres.WithInstance(conn, &postgres.Config{})
	if err != nil {
		log.Info("Error getting driver: ", zap.Error(err))
		return nil
	}

	// Check if the table 'dataurl' exists
	rows, err := conn.Query("SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'dataurl')")
	if err != nil {
		log.Info("Error checking for existing table: ", zap.Error(err))
		return nil
	}

	if err = rows.Err(); err != nil {
		log.Info("Error checking for existing table: ", zap.Error(err))
		return nil
	}

	defer rows.Close()

	var tableExists bool
	if rows.Next() {
		if err := rows.Scan(&tableExists); err != nil {
			log.Info("Error scanning for existing table: ", zap.Error(err))
			return nil
		}
	}

	if !tableExists {
		dir, err := os.Getwd()
		if err != nil {
			log.Info("Error getting getwd: ", zap.Error(err))
			return nil
		}

		// fix error test path
		var path string
		if filepath.Base(dir) == "shortener" {
			path = "../../"
		}

		m, err := migrate.NewWithDatabaseInstance(
			fmt.Sprintf("file://%smigrations", path),
			"postgres",
			driver)

		if err != nil {
			log.Info("Error creating migration instance: ", zap.Error(err))
			return nil
		}

		err = m.Up()
		if err != nil {
			log.Info("Error while performing migration: ", zap.Error(err))
			return nil
		}

		log.Info("Connected!")
	}
	return &BDKeeper{
		conn: conn,
		log:  log,
	}
}

// Load retrieves URL data from the PostgreSQL database and returns it as a map.
func (bdk *BDKeeper) Load() (storage.StorageURL, error) {
	ctx := context.Background()

	// get data from bd
	rows, err := bdk.conn.QueryContext(ctx, `SELECT correlation_id, short_url, original_url, user_id, is_deleted FROM dataurl`)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	data := make(storage.StorageURL)
	for rows.Next() {
		record := models.DataURL{}

		s := reflect.ValueOf(&record).Elem()
		numCols := s.NumField()
		columns := make([]interface{}, numCols)
		for i := 0; i < numCols; i++ {
			field := s.Field(i)
			columns[i] = field.Addr().Interface()
		}

		err = rows.Scan(columns...)
		if err != nil {
			bdk.log.Info("row scan error: ", zap.Error(err))
		}

		var u *url.URL
		u, err = url.Parse(record.ShortURL)
		if err != nil {
			panic(err)
		}

		key := u.Path
		key = strings.Replace(key, "/", "", -1)
		data[key] = record
	}

	if err = rows.Err(); err != nil {
		return data, err
	}

	return data, nil
}

// LoadUsers retrieves user data from the PostgreSQL database and returns it as a map.
func (bdk *BDKeeper) LoadUsers() (storage.StorageUser, error) {
	ctx := context.Background()

	// get data from bd
	rows, err := bdk.conn.QueryContext(ctx, `SELECT id, name, email, hash FROM users`)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	data := make(storage.StorageUser)
	for rows.Next() {
		record := models.DataUser{}

		s := reflect.ValueOf(&record).Elem()
		numCols := s.NumField()
		columns := make([]interface{}, numCols)
		for i := 0; i < numCols; i++ {
			field := s.Field(i)
			columns[i] = field.Addr().Interface()
		}

		err = rows.Scan(columns...)
		if err != nil {
			bdk.log.Info("row scan error: ", zap.Error(err))
		}
		data[record.Email] = record
	}

	if err = rows.Err(); err != nil {
		return data, err
	}

	return data, nil
}

// UpdateBatch updates the is_deleted flag for the specified URLs in the PostgreSQL database.
func (bdk *BDKeeper) UpdateBatch(data ...models.DeleteURL) error {
	ctx := context.Background()

	valueStrings := make([]string, 0, len(data))
	valueArgs := make([]interface{}, 0, len(data)*2)
	i := 0

	for _, u := range data {
		for _, k := range u.ShortURLs {
			valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d)", i*2+1, i*2+2))
			valueArgs = append(valueArgs, k)
			valueArgs = append(valueArgs, u.UserID)
			i++
		}
	}
	stmt := fmt.Sprintf(
		`WITH _data (short_url, user_id) 
		AS (VALUES %s)
		UPDATE dataurl AS d
		SET is_deleted = TRUE
		FROM _data
		WHERE d.short_url LIKE '%%' || _data.short_url || '%%'
			AND d.user_id = _data.user_id`,
		strings.Join(valueStrings, ","))
	_, err := bdk.conn.ExecContext(ctx, stmt, valueArgs...)

	if err != nil {
		return err
	}

	return nil
}

// Save inserts or updates the specified URL data in the PostgreSQL database.
// It returns the saved data along with any error encountered.
func (bdk *BDKeeper) Save(key string, data models.DataURL) (models.DataURL, error) {
	ctx := context.Background()

	var id string
	if data.UUID == "" {
		neuuid := uuid.New()
		id = neuuid.String()
	} else {
		id = data.UUID
	}
	_, err := bdk.conn.ExecContext(ctx,
		`INSERT INTO dataurl (
			correlation_id,
			short_url,
			original_url,
			user_id,
			is_deleted)
		VALUES ($1, $2, $3, $4, $5) 
		RETURNING original_url`,
		id, data.ShortURL, data.OriginalURL, data.UserID, data.DeletedFlag)

	row := bdk.conn.QueryRowContext(ctx, `
	SELECT
		d.correlation_id,
		d.short_url  ,
		d.original_url,
		d.user_id,
		d.is_deleted	 
	FROM dataurl d	 
	WHERE
		d.original_url = $1`,
		data.OriginalURL,
	)

	// read the values from the database record into the corresponding fields of the structure
	var m models.DataURL
	nerr := row.Scan(&m.UUID, &m.ShortURL, &m.OriginalURL, &m.UserID, &m.DeletedFlag)
	if nerr != nil {
		bdk.log.Info("row scan error: ", zap.Error(err))
		return data, nerr
	}

	if err != nil {
		var e *pgconn.PgError
		if errors.As(err, &e) && e.Code == pgerrcode.UniqueViolation {
			bdk.log.Info("unique field violation on column: ", zap.Error(err))

			return m, storage.ErrConflict
		}
		return m, err
	}

	return m, nil
}

// SaveUser inserts or updates the specified user data in the PostgreSQL database.
// It returns the saved data along with any error encountered.
func (bdk *BDKeeper) SaveUser(key string, data models.DataUser) (models.DataUser, error) {
	ctx := context.Background()

	var id string
	if data.UUID == "" {
		neuuid := uuid.New()
		id = neuuid.String()
	} else {
		id = data.UUID
	}

	_, err := bdk.conn.ExecContext(ctx,
		`INSERT INTO users (
			id,
			email,
			hash,
			name)
		VALUES ($1, $2, $3, $4) RETURNING id`,
		id, data.Email, data.Hash, data.Name)

	var (
		cond string
		hash []byte
	)

	if data.Hash != nil {
		cond = "AND u.hash = $2"
		hash = data.Hash
	}

	stmt := fmt.Sprintf(`
	SELECT
		u.id,
		u.email,
		u.hash,
		u.name  	 
	FROM users u	 
	WHERE
		u.email = $1 %s`, cond)
	row := bdk.conn.QueryRowContext(ctx, stmt, data.Email, hash)

	// read the values from the database record into the corresponding fields of the structure
	var m models.DataUser
	nerr := row.Scan(&m.UUID, &m.Email, &m.Hash, &m.Name)
	if nerr != nil {
		return data, nerr
	}

	if err != nil {
		var e *pgconn.PgError
		if errors.As(err, &e) && e.Code == pgerrcode.UniqueViolation {
			bdk.log.Info("unique field violation on column: ", zap.Error(err))

			return m, storage.ErrConflict
		}
		return m, err
	}

	return m, nil
}

// SaveBatch inserts or updates the specified batch of URL data in the PostgreSQL database.
// It returns any error encountered during the operation.
func (bdk *BDKeeper) SaveBatch(data storage.StorageURL) error {
	ctx := context.Background()

	valueStrings := make([]string, 0, len(data))
	valueArgs := make([]interface{}, 0, len(data)*5)
	i := 0
	for _, u := range data {
		valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d, $%d, $%d, $%d)",
			i*5+1, i*5+2, i*5+3, i*5+4, i*5+5))
		valueArgs = append(valueArgs, u.UUID)
		valueArgs = append(valueArgs, u.ShortURL)
		valueArgs = append(valueArgs, u.OriginalURL)
		valueArgs = append(valueArgs, u.UserID)
		valueArgs = append(valueArgs, u.DeletedFlag)
		i++
	}

	stmt := fmt.Sprintf(
		`INSERT INTO dataurl (
		correlation_id,
		short_url,
		original_url,
		user_id,
		is_deleted)
		VALUES %s ON CONFLICT (original_url) DO NOTHING`,
		strings.Join(valueStrings, ","))
	_, err := bdk.conn.ExecContext(ctx, stmt, valueArgs...)

	if err != nil {
		return err
	}

	return nil
}

// Ping checks the connectivity to the PostgreSQL database and returns true if successful, otherwise false.
func (bdk *BDKeeper) Ping() bool {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	if err := bdk.conn.PingContext(ctx); err != nil {
		return false
	}

	return true
}

// Close closes the connection to the PostgreSQL database and returns true if successful, otherwise false.
func (bdk *BDKeeper) Close() bool {
	bdk.log.Info("Stop database")
	bdk.conn.Close()
	bdk.log.Info("All sql queries are completed")
	return true
}
