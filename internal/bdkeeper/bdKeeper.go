package bdkeeper

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/wurt83ow/tinyurl/cmd/shortener/storage"
	"github.com/wurt83ow/tinyurl/internal/models"
)

type Log interface {
	Info(string, ...zapcore.Field)
}
type BDKeeper struct {
	conn *sql.DB
	log  Log
}

func NewBDKeeper(dsn func() string, log Log) *BDKeeper {

	addr := dsn()
	if addr == "" {
		log.Info("database dsn is empty")
		return nil
	}

	conn, err := sql.Open("pgx", dsn())
	if err != nil {
		log.Info("Unable to connection to database: ", zap.Error(err))
	}

	driver, err := postgres.WithInstance(conn, &postgres.Config{})
	if err != nil {

		log.Info("error getting driver: ", zap.Error(err))
	}

	dir, err := os.Getwd()
	if err != nil {
		log.Info("error getting getwd: ", zap.Error(err))
	}

	// fix error test path
	path := ""
	if filepath.Base(dir) == "tinyurl" {
		path = "cmd/shortener/"
	}

	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s/%smigrations", dir, path),
		"postgres",
		driver)

	if err != nil {
		log.Info("Error creating migration instance : ", zap.Error(err))
	}
	err = m.Up()
	if err != nil {
		log.Info("Error while performing migration: ", zap.Error(err))
	}

	log.Info("Connected!")
	return &BDKeeper{
		conn: conn,
		log:  log,
	}
}

func (bdk *BDKeeper) Load() (storage.StorageURL, error) {

	ctx := context.Background()
	data := make(storage.StorageURL)
	// запрашиваем данные обо всех сообщениях пользователя, без самого текста
	rows, err := bdk.conn.QueryContext(ctx, `SELECT correlation_id, short_url, original_url, user_id FROM dataurl`)

	if err != nil {
		return nil, err
	}

	// не забываем закрыть курсор после завершения работы с данными
	defer rows.Close()

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
	if err = rows.Err(); err != nil {
		return data, err
	}

	return data, nil
}

// LoadUsers implements storage.Keeper.
func (bdk *BDKeeper) LoadUsers() (storage.StorageUser, error) {

	ctx := context.Background()
	data := make(storage.StorageUser)

	// запрашиваем данные обо всех сообщениях пользователя, без самого текста
	rows, err := bdk.conn.QueryContext(ctx, `SELECT id, name, email, hash FROM users`)

	if err != nil {
		return nil, err
	}

	// не забываем закрыть курсор после завершения работы с данными
	defer rows.Close()

	for rows.Next() {
		record := models.DataUser{}

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
		data[record.Email] = record
	}
	if err = rows.Err(); err != nil {

		return data, err
	}

	return data, nil
}

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
		"INSERT INTO dataurl (correlation_id, short_url, original_url, user_id) VALUES ($1, $2, $3, $4) RETURNING original_url",
		id, data.ShortURL, data.OriginalURL, data.UserID)

	row := bdk.conn.QueryRowContext(ctx, `
	SELECT
		d.correlation_id,
		d.short_url  ,
		d.original_url,
		d.user_id		 
	FROM dataurl d	 
	WHERE
		d.original_url = $1
`,
		data.OriginalURL,
	)

	// считываем значения из записи БД в соответствующие поля структуры
	var m models.DataURL
	nerr := row.Scan(&m.UUID, &m.ShortURL, &m.OriginalURL, &m.UserID)
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

// func (bdk *BDKeeper) CreateDefaultUser() (string, error) {
// 	ctx := context.Background()

// 	id := "00000000-0000-0000-0000-000000000000"
// 	_, err := bdk.conn.ExecContext(ctx,
// 		"INSERT INTO users (id, email, hash, name) VALUES ($1, $2, $3, $4) RETURNING id",
// 		id, "default", "default", "default")

// 	if err != nil {
// 		return "", err
// 	}
// 	return id, nil
// }

// comment
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
		"INSERT INTO users (id, email, hash, name) VALUES ($1, $2, $3, $4) RETURNING id",
		id, data.Email, data.Hash, data.Name)

	row := bdk.conn.QueryRowContext(ctx, `
	SELECT
		u.id,
		u.email,
		u.hash,
		u.name  	 
	FROM users u	 
	WHERE
		u.email = $1 AND u.hash = $2
`,
		data.Email, data.Hash,
	)

	// считываем значения из записи БД в соответствующие поля структуры
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

func (bdk *BDKeeper) SaveBatch(data storage.StorageURL) error {
	ctx := context.Background()
	// func BulkInsert(unsavedRows []*ExampleRowStruct) error {
	valueStrings := make([]string, 0, len(data))
	valueArgs := make([]interface{}, 0, len(data)*4)
	i := 0
	for _, post := range data {
		valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d, $%d, $%d)", i*4+1, i*4+2, i*4+3, i*4+4))
		valueArgs = append(valueArgs, post.UUID)
		valueArgs = append(valueArgs, post.ShortURL)
		valueArgs = append(valueArgs, post.OriginalURL)
		valueArgs = append(valueArgs, post.UserID)
		i++
	}
	stmt := fmt.Sprintf("INSERT INTO dataurl (correlation_id, short_url, original_url, user_id) VALUES %s ON CONFLICT (original_url) DO NOTHING",
		strings.Join(valueStrings, ","))
	_, err := bdk.conn.ExecContext(ctx, stmt, valueArgs...)
	if err != nil {
		fmt.Println("978798789sd7f9s87df8s7df9d8s7f9s8f", err)
		return err
	}

	return nil

}

func (bdk *BDKeeper) Ping() bool {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	if err := bdk.conn.PingContext(ctx); err != nil {
		return false
	}

	return true
}

func (bdk *BDKeeper) Close() bool {
	bdk.conn.Close()
	return true
}
