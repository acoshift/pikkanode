package auth

import (
	"context"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/acoshift/pikkanode/internal/password"
	"github.com/acoshift/pikkanode/internal/session"
	"github.com/acoshift/pikkanode/internal/validator"
)

type SignUpRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

var reUsername = regexp.MustCompile(`^[a-zA-Z0-9]*$`)

func (req *SignUpRequest) Valid() error {
	v := validator.New()
	v.Must(req.Username != "", "username required")
	v.Must(utf8.RuneCountInString(req.Username) >= 6, "username minimum 6 characters")
	v.Must(utf8.RuneCountInString(req.Username) <= 15, "username maximum 15 characters")
	v.Must(reUsername.MatchString(req.Username), "invalid username")

	v.Must(req.Password != "", "password required")
	v.Must(utf8.RuneCountInString(req.Password) >= 6, "password minimum 6 characters")
	v.Must(utf8.RuneCountInString(req.Password) <= 500, "password maximum 500 characters")

	return v.Error()
}

func SignUp(ctx context.Context, req *SignUpRequest) (*struct{}, error) {
	hashed, err := password.Hash(req.Password)
	if err != nil {
		return nil, err
	}

	_, err = insertUser(ctx, req.Username, hashed)
	if err == errUsernameDuplicated {
		return nil, errUsernameNotAvailable
	}
	if err != nil {
		return nil, err
	}

	return new(struct{}), nil
}

type SignInRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (req *SignInRequest) Valid() error {
	v := validator.New()
	v.Must(req.Username != "", "username required")
	v.Must(utf8.RuneCountInString(req.Username) <= 15, "username maximum 15 characters")

	v.Must(req.Password != "", "password required")
	v.Must(utf8.RuneCountInString(req.Password) <= 500, "password maximum 500 characters")

	return v.Error()
}

func SignIn(ctx context.Context, req *SignInRequest) (*struct{}, error) {
	userID, hashed, err := getUserIDAndPasswordByUsername(ctx, req.Username)
	if err == errNotFound {
		return nil, errInvalidCredentials
	}

	if !password.Compare(hashed, req.Password) {
		return nil, errInvalidCredentials
	}

	session.Get(ctx).Set("user_id", userID)

	return new(struct{}), nil
}

func SignOut(ctx context.Context, _ *struct{}) (*struct{}, error) {
	session.Get(ctx).Destroy()
	return new(struct{}), nil
}

type CheckRequest struct{}

func (*CheckRequest) AdaptRequest(r *http.Request) {
	if r.Method == http.MethodGet {
		r.Method = http.MethodPost
		r.Header.Set("Content-Type", "application/json")
		r.Body = ioutil.NopCloser(strings.NewReader("{}"))
	}
}

type CheckResult struct {
	OK bool `json:"ok"`
}

func Check(ctx context.Context, _ *CheckRequest) (*CheckResult, error) {
	userID := session.GetUserID(ctx)

	ok := userID != ""

	return &CheckResult{OK: ok}, nil
}
