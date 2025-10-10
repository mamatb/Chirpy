package web

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/mamatb/Chirpy/internal/auth"
	"github.com/mamatb/Chirpy/internal/database"
)

func HandlerGetApiHealth() func(http.ResponseWriter, *http.Request) {
	return respPlainOk
}

func HandlerApp(config *ApiConfig) http.Handler {
	return config.middleMetricsInc(http.StripPrefix(
		"/app/",
		http.FileServer(http.Dir(".")),
	))
}

func HandlerGetAdminMetrics(config *ApiConfig) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set(HeaderContentType, ContentTypeHtml)
		if _, err := w.Write([]byte(fmt.Sprintf(""+
			"<html>\n"+
			"  <body>\n"+
			"    <h1>Welcome, Chirpy Admin</h1>\n"+
			"    <p>Chirpy has been visited %d times!</p>\n"+
			"  </body>\n"+
			"</html>\n",
			config.FileserverHits.Load(),
		))); err != nil {
			log.Fatal(err)
		}
	}
}

func HandlerPostAdminReset(config *ApiConfig) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if config.Platform != "dev" {
			respPlainForbidden(w, r)
			return
		}
		if config.DBQueries.DeleteUsers(r.Context()) != nil {
			respPlainError(w, r, ErrorSomethingWentWrong)
			return
		}
		config.middleMetricsReset(respPlainOk)
	}
}

func HandlerPostApiUsers(config *ApiConfig) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error
		var hash string
		var user database.User
		request := struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}{}
		if json.NewDecoder(r.Body).Decode(&request) != nil {
			respJsonError(w, r, ErrorSomethingWentWrong)
			return
		}
		if hash, err = auth.HashPassword(request.Password); err != nil {
			respJsonError(w, r, ErrorSomethingWentWrong)
			return
		}
		if user, err = config.DBQueries.CreateUser(
			r.Context(),
			database.CreateUserParams{
				Email:          request.Email,
				HashedPassword: hash,
			},
		); err != nil {
			respJsonError(w, r, ErrorSomethingWentWrong)
			return
		}
		respJsonUserCreated(w, r, user)
	}
}

func HandlerPostApiLogin(config *ApiConfig) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error
		var token, refreshToken string
		var user database.User
		request := struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}{}
		if json.NewDecoder(r.Body).Decode(&request) != nil {
			respJsonError(w, r, ErrorSomethingWentWrong)
			return
		}
		if user, err = config.DBQueries.GetUser(
			r.Context(),
			request.Email,
		); err != nil {
			respJsonError(w, r, ErrorSomethingWentWrong)
			return
		}
		if auth.CheckPasswordHash(request.Password, user.HashedPassword) != nil {
			respPlainUnauthorized(w, r, "Incorrect email or password")
			return
		}
		if token, err = auth.MakeJWT(
			user.ID,
			config.Secret,
			time.Hour,
		); err != nil {
			respJsonError(w, r, ErrorSomethingWentWrong)
			return
		}
		if refreshToken, err = auth.MakeRefreshToken(); err != nil {
			respJsonError(w, r, ErrorSomethingWentWrong)
			return
		}
		if _, err = config.DBQueries.CreateRefreshToken(
			r.Context(),
			database.CreateRefreshTokenParams{
				Token:     refreshToken,
				UserID:    uuid.NullUUID{UUID: user.ID, Valid: true},
				ExpiresAt: time.Now().Add(time.Hour * 24 * 60),
			},
		); err != nil {
			respJsonError(w, r, ErrorSomethingWentWrong)
			return
		}
		respJsonUser(w, r, user, token, refreshToken)
	}
}

func HandlerPostApiRefresh(config *ApiConfig) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error
		var token, refreshToken string
		var user database.User
		if refreshToken, err = auth.GetBearerToken(r.Header); err != nil {
			respPlainUnauthorized(w, r, "Missing refresh token")
			return
		}
		if user, err = config.DBQueries.GetUserFromRefreshToken(
			r.Context(),
			refreshToken,
		); err != nil || user.ID == uuid.Nil {
			respPlainUnauthorized(w, r, "Invalid or expired refresh token")
			return
		}
		if token, err = auth.MakeJWT(
			user.ID,
			config.Secret,
			time.Hour,
		); err != nil {
			respJsonError(w, r, ErrorSomethingWentWrong)
			return
		}
		respJsonToken(w, r, token)
	}
}

func HandlerPostApiRevoke(config *ApiConfig) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error
		var refreshToken string
		if refreshToken, err = auth.GetBearerToken(r.Header); err != nil {
			respPlainError(w, r, "Missing refresh token")
			return
		}
		if config.DBQueries.DeleteRefreshToken(
			r.Context(),
			refreshToken,
		) != nil {
			respPlainError(w, r, ErrorSomethingWentWrong)
			return
		}
		w.WriteHeader(204)
	}
}

func HandlerGetApiChirpsId(config *ApiConfig) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error
		var chirp database.Chirp
		var chirpId uuid.UUID
		if chirpId, err = uuid.Parse(r.PathValue("id")); err != nil {
			respJsonError(w, r, ErrorSomethingWentWrong)
			return
		}
		if chirp, err = config.DBQueries.GetChirp(
			r.Context(),
			chirpId,
		); err != nil {
			respJsonError(w, r, ErrorSomethingWentWrong)
			return
		}
		if chirp.ID == uuid.Nil {
			respPlainNotFound(w, r)
			return
		}
		respJsonChirp(w, r, chirp)
	}
}

func HandlerGetApiChirps(config *ApiConfig) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error
		var chirps []database.Chirp
		if chirps, err = config.DBQueries.GetChirps(r.Context()); err != nil {
			respJsonError(w, r, ErrorSomethingWentWrong)
			return
		}
		respJsonChirps(w, r, chirps)
	}
}

func HandlerPostApiChirps(config *ApiConfig,
	profanities map[string]bool) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error
		var token string
		var userId uuid.UUID
		var chirp database.Chirp
		if token, err = auth.GetBearerToken(r.Header); err != nil {
			respPlainUnauthorized(w, r, "Missing token")
			return
		}
		if userId, err = auth.ValidateJWT(token, config.Secret); err != nil {
			respPlainUnauthorized(w, r, "Invalid token")
			return
		}
		request := struct {
			Body string `json:"body"`
		}{}
		if json.NewDecoder(r.Body).Decode(&request) != nil {
			respJsonError(w, r, ErrorSomethingWentWrong)
			return
		}
		if len(request.Body) > 140 {
			respJsonError(w, r, "Chirp is too long")
			return
		}
		if chirp, err = config.DBQueries.CreateChirp(
			r.Context(),
			database.CreateChirpParams{
				Body:   cleanProfanities(request.Body, profanities),
				UserID: uuid.NullUUID{UUID: userId, Valid: true},
			},
		); err != nil {
			respJsonError(w, r, ErrorSomethingWentWrong)
			return
		}
		respJsonChirpCreated(w, r, chirp)
	}
}
