package database

import (
	"context"
	"fmt"

	botCMD "github.com/Ummuys/TWB/bot"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

func Connect(connect string) (*pgxpool.Pool, context.Context, error) {
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, connect)
	if err != nil {
		return nil, nil, fmt.Errorf("can't open pool: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, nil, fmt.Errorf("no answer from pool: %w", err)
	}

	return pool, ctx, err
}

func CreateItems(logger *zap.Logger, pool *pgxpool.Pool, ctx context.Context) error {
	queryCreateSchema := `
	CREATE SCHEMA IF NOT EXISTS twb;
	`
	queryCreateTable := `
	CREATE TABLE IF NOT exists twb.user_tg_info(
	id bigint primary key,
	user_state varchar(10) default '',
	fav_city1 varchar(20) default '',
	fav_city2 varchar(20) default '',
	fav_city3 varchar(20) default ''
	)
	`
	_, err := pool.Exec(ctx, queryCreateSchema)
	if err != nil {
		return fmt.Errorf("can't create schema: %w", err)
	}
	logger.Info("Schema connected")

	_, err = pool.Exec(ctx, queryCreateTable)
	if err != nil {
		return fmt.Errorf("can't create table: %w", err)
	}
	logger.Info("Table connected")

	return nil
}

func GetInfoFromTable(logger *zap.Logger, pool *pgxpool.Pool, ctx context.Context, usersInfo map[int64]botCMD.ChatInfo) error {
	// Data from table
	query := `SELECT id, user_state, fav_city1, fav_city2, fav_city3 FROM twb.user_tg_info`

	rows, err := pool.Query(ctx, query)
	if err != nil {
		return fmt.Errorf("error select: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var chatInfo botCMD.ChatInfo
		err := rows.Scan(&chatInfo.Id, &chatInfo.UserState, &chatInfo.FavoriteCitys[0], &chatInfo.FavoriteCitys[1], &chatInfo.FavoriteCitys[2])
		if err != nil {
			return fmt.Errorf("error Scan: %w", err)
		}
		usersInfo[chatInfo.Id] = chatInfo
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("error iteration rows: %w", err)
	}
	logger.Info("Download info from table is successful")

	return nil
}

func FillInfoFromMap(ctx context.Context, pool *pgxpool.Pool, data map[int64]botCMD.ChatInfo) error {
	batch := &pgx.Batch{}

	for _, chatInfo := range data {
		sql := `
			INSERT INTO twb.user_tg_info (id, user_state, fav_city1, fav_city2, fav_city3)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (id) DO UPDATE
			SET 
				user_state = EXCLUDED.user_state,
				fav_city1  = EXCLUDED.fav_city1,
				fav_city2  = EXCLUDED.fav_city2,
				fav_city3  = EXCLUDED.fav_city3;
		`

		batch.Queue(sql,
			chatInfo.Id,
			chatInfo.UserState,
			chatInfo.FavoriteCitys[0],
			chatInfo.FavoriteCitys[1],
			chatInfo.FavoriteCitys[2],
		)
	}

	br := pool.SendBatch(ctx, batch)
	defer br.Close()

	for i := 0; i < len(data); i++ {
		_, err := br.Exec()
		if err != nil {
			return fmt.Errorf("error upsert #%d: %w", i, err)
		}
	}

	return nil
}
