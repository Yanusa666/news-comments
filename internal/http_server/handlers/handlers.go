package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"github.com/rs/zerolog"
	"net/http"
	"sf-news-comments/internal/comments"
	"sf-news-comments/internal/config"
	"sf-news-comments/internal/constants"
	"strconv"
)

type Handler struct {
	cfg      *config.Config
	lgr      zerolog.Logger
	comments *comments.Comments
}

func NewHandler(cfg *config.Config, lgr zerolog.Logger, comments *comments.Comments) *Handler {
	return &Handler{
		cfg:      cfg,
		lgr:      lgr,
		comments: comments,
	}
}

func (h *Handler) AddComment(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()

	request := new(AddCommentReq)
	err := decoder.Decode(request)
	if err != nil {
		resp, _ := json.Marshal(ErrorResp{Error: fmt.Sprintf("incorrect request: %s", err.Error())})
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, string(resp))
		return
	}

	ctx := r.Context()
	requestId, _ := ctx.Value(constants.RequestIdKey).(string)

	lgr := h.lgr.With().
		Str("handler", "AddComment").
		Str(constants.RequestIdKey, requestId).
		Dict("request", zerolog.Dict().
			Uint64("new_id", request.NewId).
			Interface("parent_id", request.ParentId).
			Str("text", request.Text)).
		Logger()

	err = h.comments.Model.Add(request.NewId, request.ParentId, request.Text)
	if err != nil {
		resp, _ := json.Marshal(ErrorResp{Error: fmt.Sprintf("internal error: %s", err.Error())})
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, string(resp))
		return
	}

	lgr.Debug().Msg("executed")

	resp, _ := json.Marshal(AddCommentResp{Status: "success"})
	fmt.Fprintf(w, string(resp))
}

func (h *Handler) GetComments(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	ctx := r.Context()
	requestId, _ := ctx.Value(constants.RequestIdKey).(string)

	newIdStr := ps.ByName("new_id")
	lgr := h.lgr.With().
		Str("handler", "GetComments").
		Str(constants.RequestIdKey, requestId).
		Dict("request", zerolog.Dict().
			Str("new_id", newIdStr)).
		Logger()

	newId, err := strconv.ParseUint(newIdStr, 10, 64)
	if err != nil {
		resp, _ := json.Marshal(ErrorResp{Error: fmt.Sprintf("incorrect count: %s", newIdStr)})
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, string(resp))
		return
	}

	commentsList, err := h.comments.Model.GetList(newId)
	if err != nil {
		resp, _ := json.Marshal(ErrorResp{Error: fmt.Sprintf("internal error: %s", err.Error())})
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, string(resp))
		return
	}

	comments := make([]CommentEntity, 0, 10)
	for _, i := range commentsList {
		comments = append(comments, CommentEntity{
			Id:       i.Id,
			ParentId: i.ParentId,
			Text:     i.Text,
			PubDate:  JsonTime(i.PubDate),
		})
	}

	lgr.Debug().Msg("executed")

	resp, _ := json.Marshal(GetCommentsResp{Comments: comments})
	fmt.Fprintf(w, string(resp))
}
