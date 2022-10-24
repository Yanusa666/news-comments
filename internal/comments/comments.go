package comments

import (
	"github.com/rs/zerolog"
	"sf-news-comments/internal/config"
)

type Comments struct {
	cfg   *config.Config
	lgr   zerolog.Logger
	Model *Model
}

func NewComments(cfg *config.Config, lgr zerolog.Logger) *Comments {
	model := NewModel(cfg, lgr)

	n := &Comments{
		cfg:   cfg,
		lgr:   lgr,
		Model: model,
	}

	return n
}

func (n *Comments) Shutdown() error {
	n.Model.connPool.Close()
	return nil
}
