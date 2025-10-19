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

func HandlerGetApiHealth() http.HandlerFunc {
	return respPlainOk
}

func HandlerApp(config *ApiConfig) http.Handler {
	return config.middleMetricsInc(http.StripPrefix(
		"/app/",
		http.FileServer(http.Dir(".")),
	))
}

func HandlerGetAdminMetrics(config *ApiConfig) http.HandlerFunc {
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

func HandlerPostAdminReset(config *ApiConfig) http.HandlerFunc {
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

func HandlerPostApiUsers(config *ApiConfig) http.HandlerFunc {
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

func HandlerPutApiUsers(config *ApiConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error
		var token, hash string
		var userId uuid.UUID
		var user database.User
		if token, err = auth.GetBearerToken(r.Header); err != nil {
			respJsonUnauthorized(w, r, ErrorMissingToken)
			return
		}
		if userId, err = auth.ValidateJWT(token, config.Secret); err != nil {
			respJsonUnauthorized(w, r, ErrorInvalidToken)
			return
		}
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
		if user, err = config.DBQueries.UpdateUserCredentials(
			r.Context(),
			database.UpdateUserCredentialsParams{
				ID:             userId,
				Email:          request.Email,
				HashedPassword: hash,
			},
		); err != nil {
			respJsonError(w, r, ErrorSomethingWentWrong)
			return
		}
		respJsonUser(w, r, user, "", "")
	}
}

func HandlerPostApiLogin(config *ApiConfig) http.HandlerFunc {
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
		); err != nil || auth.ValidateHash(request.Password, user.HashedPassword) != nil {
			respJsonUnauthorized(w, r, "Incorrect email or password")
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

func HandlerPostApiRefresh(config *ApiConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error
		var token, refreshToken string
		var user database.User
		if refreshToken, err = auth.GetBearerToken(r.Header); err != nil {
			respJsonUnauthorized(w, r, ErrorMissingRefreshToken)
			return
		}
		if user, err = config.DBQueries.GetUserFromRefreshToken(
			r.Context(),
			refreshToken,
		); err != nil || user.ID == uuid.Nil {
			respJsonUnauthorized(w, r, ErrorInvalidRefreshToken)
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

func HandlerPostApiRevoke(config *ApiConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error
		var refreshToken string
		if refreshToken, err = auth.GetBearerToken(r.Header); err != nil {
			respPlainError(w, r, ErrorMissingRefreshToken)
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

func HandlerGetApiChirpsId(config *ApiConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error
		var chirpId uuid.UUID
		var chirp database.Chirp
		if chirpId, err = uuid.Parse(r.PathValue("id")); err != nil {
			respJsonError(w, r, ErrorSomethingWentWrong)
			return
		}
		if chirp, err = config.DBQueries.GetChirp(
			r.Context(),
			chirpId,
		); err != nil || chirp.ID == uuid.Nil {
			respPlainNotFound(w, r)
			return
		}
		respJsonChirp(w, r, chirp)
	}
}

func HandlerGetApiChirps(config *ApiConfig) http.HandlerFunc {
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

func HandlerPostApiChirps(config *ApiConfig, profanities map[string]bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error
		var token string
		var userId uuid.UUID
		var chirp database.Chirp
		if token, err = auth.GetBearerToken(r.Header); err != nil {
			respJsonUnauthorized(w, r, ErrorMissingToken)
			return
		}
		if userId, err = auth.ValidateJWT(token, config.Secret); err != nil {
			respJsonUnauthorized(w, r, ErrorInvalidToken)
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

func HandlerDeleteApiChirpsId(config *ApiConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error
		var token string
		var userId, chirpId uuid.UUID
		var chirp database.Chirp
		if token, err = auth.GetBearerToken(r.Header); err != nil {
			respPlainUnauthorized(w, r)
			return
		}
		if userId, err = auth.ValidateJWT(token, config.Secret); err != nil {
			respPlainUnauthorized(w, r)
			return
		}
		if chirpId, err = uuid.Parse(r.PathValue("id")); err != nil {
			respPlainError(w, r, ErrorSomethingWentWrong)
			return
		}
		if chirp, err = config.DBQueries.GetChirp(
			r.Context(),
			chirpId,
		); err != nil || chirp.ID == uuid.Nil {
			respPlainNotFound(w, r)
			return
		}
		if chirp.UserID.UUID != userId {
			respPlainForbidden(w, r)
			return
		}
		if config.DBQueries.DeleteChirp(
			r.Context(),
			chirp.ID,
		) != nil {
			respPlainError(w, r, ErrorSomethingWentWrong)
			return
		}
		w.WriteHeader(204)
	}
}

func HandlerPostApiPolkaWebhooks(config *ApiConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error
		var apiKey string
		var user database.User
		if apiKey, err = auth.GetApiKey(r.Header); err != nil || apiKey != config.PolkaKey {
			respPlainUnauthorized(w, r)
			return
		}
		request := struct {
			Event string `json:"event"`
			Data  struct {
				UserId uuid.UUID `json:"user_id"`
			} `json:"data"`
		}{}
		if json.NewDecoder(r.Body).Decode(&request) != nil {
			respPlainError(w, r, ErrorSomethingWentWrong)
			return
		}
		if request.Event != "user.upgraded" {
			w.WriteHeader(204)
			return
		}
		if user, err = config.DBQueries.UpdateUserRed(
			r.Context(),
			request.Data.UserId,
		); err != nil || user.ID == uuid.Nil {
			respPlainNotFound(w, r)
			return
		}
		w.WriteHeader(204)
	}
}
