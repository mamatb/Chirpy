package web

import (
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/mamatb/Chirpy/database"
)

const (
	daysInMonth = 30
	hoursInDay  = 24

	contentTypeHtml           = "text/html; charset=utf-8"
	contentTypeJson           = "application/json; charset=utf-8"
	contentTypePlain          = "text/plain; charset=utf-8"
	cwd                       = "."
	empty                     = ""
	platformDev               = "dev"
	errorChirpTooLong         = "Chirp is too long"
	errorInvalidEmailPassword = "Invalid email or password"
	errorInvalidToken         = "Invalid token"
	errorInvalidRefreshToken  = "Invalid refresh token"
	errorMissingToken         = "Missing token"
	errorMissingRefreshToken  = "Missing refresh token"
	errorSomethingWentWrong   = "Something went wrong"
	headerContentType         = "Content-Type"
	httpForbiddenPlain        = "FORBIDDEN"
	httpNotFoundPlain         = "NOT FOUND"
	httpOkPlain               = "OK"
	httpUnauthorizedPlain     = "UNAUTHORIZED"
	orderDesc                 = "desc"
	polkaEventUserUpgraded    = "user.upgraded"
	profanitiesReplacement    = "****"
	space                     = " "
)

type ApiConfig struct {
	Platform       string
	PolkaKey       string
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
	IsChirpyRed  bool      `json:"is_chirpy_red"`
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
