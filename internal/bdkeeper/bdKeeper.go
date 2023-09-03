package bdkeeper

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap/zapcore"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type Log interface {
	Info(string, ...zapcore.Field)
}
type BDKeeper struct {
	pool *pgxpool.Pool
	log  Log
}

func NewBDKeeper(pool *pgxpool.Pool, log Log) *BDKeeper {
	return &BDKeeper{
		pool: pool,
		log:  log,
	}
}

func (bdk *BDKeeper) Ping() bool {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	if err := bdk.pool.Ping(ctx); err != nil {
		fmt.Println(err)
		return false
	}
	fmt.Println("77777777777777777777777777777")
	return true
}
