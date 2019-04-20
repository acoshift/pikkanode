package user

import (
	"context"
	"database/sql"
	"net/http"
	"net/url"
	"unicode/utf8"

	"github.com/acoshift/arpc"
	"github.com/acoshift/pgsql/pgctx"

	"github.com/acoshift/pikkanode/internal/file"
	"github.com/acoshift/pikkanode/internal/session"
	"github.com/acoshift/pikkanode/internal/validator"
)

var (
	errUserNotFound = arpc.NewError("user not found")
	errFollowSelf   = arpc.NewError("can not follow self")
)

type ProfileRequest struct {
	Username string `json:"username"`
}

func (*ProfileRequest) AdaptRequest(r *http.Request) {
	if r.Method == http.MethodGet {
		r.ParseForm()
		r.Method = http.MethodPost
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		r.PostForm = r.Form
	}
}

func (req *ProfileRequest) UnmarshalForm(v url.Values) error {
	req.Username = v.Get("username")
	return nil
}

func (req *ProfileRequest) Valid() error {
	v := validator.New()
	v.Must(req.Username != "", "username required")
	v.Must(utf8.RuneCountInString(req.Username) <= 15, "username maximum 15 characters")

	return v.Error()
}

type ProfileResult struct {
	Username  string           `json:"username"`
	Photo     file.DownloadURL `json:"photo"`
	Following bool             `json:"following"`
	Follower  bool             `json:"follower"`
}

func Profile(ctx context.Context, req *ProfileRequest) (*ProfileResult, error) {
	userID := session.GetUserID(ctx)

	var r ProfileResult
	// language=SQL
	err := pgctx.QueryRow(ctx, `
		select username, photo,
		       exists(select 1 from follows where following_id = users.id and user_id = $2),
		       exists(select 1 from follows where user_id = users.id and following_id = $2)
		from users
		where username = $1
	`, req.Username, userID).Scan(
		&r.Username, &r.Photo, &r.Following, &r.Follower,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &r, nil
}

type FollowRequest struct {
	Username string `json:"username"`
	Follow   bool   `json:"follow"`
}

func (req *FollowRequest) Valid() error {
	v := validator.New()
	v.Must(req.Username != "", "username required")
	v.Must(utf8.RuneCountInString(req.Username) <= 15, "username maximum 15 characters")

	return v.Error()
}

func Follow(ctx context.Context, req *FollowRequest) (*struct{}, error) {
	userID := session.GetUserID(ctx)
	if userID == "" {
		return nil, errInvalidCredentials
	}

	followingID, err := getUserIDFromUsername(ctx, req.Username)
	if err == errNotFound {
		return nil, errUserNotFound
	}
	if err != nil {
		return nil, err
	}
	if userID == followingID {
		return nil, errFollowSelf
	}

	if req.Follow {
		// language=SQL
		_, err := pgctx.Exec(ctx, `
			insert into follows
				(user_id, following_id)
			values
				($1, $2)
			on conflict (user_id, following_id) do nothing
		`, userID, followingID)
		if err != nil {
			return &struct{}{}, err
		}
	} else {
		// language=SQL
		_, err := pgctx.Exec(ctx, `
			delete from follows
			where user_id = $1 and following_id = $2
		`, userID, followingID)
		if err != nil {
			return &struct{}{}, err
		}
	}

	return new(struct{}), nil
}
