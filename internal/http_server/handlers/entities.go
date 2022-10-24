package handlers

import (
	"fmt"
	"time"
)

type JsonTime time.Time

func (t JsonTime) MarshalJSON() (b []byte, err error) {
	tm := time.Time(t)

	return []byte(fmt.Sprintf(`"%s"`, tm.Format(time.RFC3339))), nil
}

type ErrorResp struct {
	Error string `json:"error"`
}

type AddCommentReq struct {
	NewId    uint64  `json:"new_id"`
	ParentId *uint64 `json:"parent_id"`
	Text     string  `json:"text"`
}

type AddCommentResp struct {
	Status string `json:"status"`
}

type CommentEntity struct {
	Id       uint64   `json:"id"`
	ParentId *uint64  `json:"parent_id"`
	Text     string   `json:"text"`
	PubDate  JsonTime `json:"pub_date"`
}

type GetCommentsResp struct {
	Comments []CommentEntity `json:"comments"`
}
