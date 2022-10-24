package comments

import (
	"context"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/rs/zerolog"
	"sf-news-comments/internal/config"
	"time"
)

type Model struct {
	cfg      *config.Config
	lgr      zerolog.Logger
	connPool *pgxpool.Pool
}

func NewModel(cfg *config.Config, lgr zerolog.Logger) *Model {
	lgr = lgr.With().Str("model", "comments").Logger()

	pgConf, err := pgxpool.ParseConfig(cfg.Postgres.URI)
	if err != nil {
		lgr.Fatal().Err(err).Msg("failed parse PostgreSQL config")
	}

	pgPool, err := pgxpool.ConnectConfig(context.Background(), pgConf)
	if err != nil {
		lgr.Fatal().Err(err).Msg("failed connect to PostgreSQL")
	}

	err = pgPool.Ping(context.Background())
	if err != nil {
		lgr.Fatal().Err(err).Msg("unsuccessful ping attempt")
	}

	return &Model{
		cfg:      cfg,
		lgr:      lgr,
		connPool: pgPool,
	}
}

func (m *Model) Add(newId uint64, parentId *uint64, text string) error {
	lgr := m.lgr.With().
		Str("method", "Add").
		Dict("request", zerolog.Dict().
			Uint64("newId", newId).
			Interface("parentId", parentId).
			Str("text", text)).
		Logger()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.connPool.Exec(ctx,
		`INSERT INTO comments.comments(new_id, parent_id, text) 
			 VALUES ($1,$2,$3)`, newId, parentId, text)
	if err != nil {
		lgr.Error().Err(err).Msg("insert comment to db failed")
		return err
	}

	lgr.Debug().Msg("add comment to db")

	return nil
}

func (m *Model) GetList(newId uint64) ([]CommentDb, error) {
	lgr := m.lgr.With().Uint64("new_id", newId).Logger()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.connPool.Query(ctx,
		`SELECT id, parent_id, text, pub_date 
			 FROM comments.comments
			 WHERE new_id = $1
			 ORDER BY pub_date DESC`, newId)
	if err != nil {
		lgr.Error().Err(err).Msg("select comments from db failed")
		return nil, err
	}

	comments := make([]CommentDb, 0, 4)
	for rows.Next() {
		item := CommentDb{}
		err = rows.Scan(&(item.Id), &(item.ParentId), &(item.Text), &(item.PubDate))
		if err != nil {
			lgr.Error().Err(err).Msg("scan from db failed")
			return nil, err
		}
		comments = append(comments, item)
	}

	lgr.Debug().Msg("get comments from db")

	return comments, nil
}
