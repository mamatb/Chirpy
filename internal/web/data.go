package web

import (
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/mamatb/Chirpy/internal/database"
)

const (
	HeaderContentType        = "Content-Type"
	ContentTypePlain         = "text/plain; charset=utf-8"
	ContentTypeHtml          = "text/html; charset=utf-8"
	ContentTypeJson          = "application/json; charset=utf-8"
	ErrorSomethingWentWrong  = "Something went wrong"
	ErrorMissingToken        = "Missing token"
	ErrorInvalidToken        = "Invalid token"
	ErrorMissingRefreshToken = "Missing refresh token"
	ErrorInvalidRefreshToken = "Invalid refresh token"
)

type ApiConfig struct {
	Platform       string
	Secret         string
	DBQueries      *database.Queries
	FileserverHits atomic.Int32
}

type jsonError struct {
	Error string `json:"error"`
}

type jsonUser struct {
	Id           uuid.UUID `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Email        string    `json:"email"`
	Token        string    `json:"token,omitempty"`
	RefreshToken string    `json:"refresh_token,omitempty"`
}

type jsonToken struct {
	Token string `json:"token"`
}

type jsonChirp struct {
	Id        uuid.UUID     `json:"id"`
	CreatedAt time.Time     `json:"created_at"`
	UpdatedAt time.Time     `json:"updated_at"`
	Body      string        `json:"body"`
	UserId    uuid.NullUUID `json:"user_id"`
}
