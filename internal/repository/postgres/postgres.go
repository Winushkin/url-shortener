package repository

import (
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Postgres interface {
}

type postgres struct {
	pool *pgxpool.Pool
}

func NewPostgres(pool *pgxpool.Pool) Postgres {
	return postgres{pool: pool}
}

func (p *postgres) InsertUrl() {

}
