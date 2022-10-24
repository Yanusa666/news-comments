package comments

import (
	"time"
)

type CommentDb struct {
	Id       uint64    `db:"id"`
	ParentId *uint64   `db:"parent_id"`
	Text     string    `db:"text"`
	PubDate  time.Time `db:"pub_date"`
}
