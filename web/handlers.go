package web

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"slices"
	"time"

	"github.com/google/uuid"
	"github.com/mamatb/Chirpy/auth"
	"github.com/mamatb/Chirpy/database"
)

func HandlerGetApiHealth() http.HandlerFunc {
	return respPlainOk
}

func HandlerApp(config *ApiConfig) http.Handler {
	return config.middleMetricsInc(http.StripPrefix(
		"/app/",
		http.FileServer(http.Dir(cwd)),
	))
}

func HandlerGetAdminMetrics(config *ApiConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set(headerContentType, contentTypeHtml)
		if _, err := w.Write([]byte(fmt.Sprintf(empty+
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
		if config.Platform != platformDev {
			respPlainForbidden(w, r)
			return
		}
		if config.DBQueries.DeleteUsers(r.Context()) != nil {
			respPlainBadRequest(w, r, errorSomethingWentWrong)
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
			respJsonBadRequest(w, r, errorSomethingWentWrong)
			return
		}
		if hash, err = auth.HashPassword(request.Password); err != nil {
			respJsonBadRequest(w, r, errorSomethingWentWrong)
			return
		}
		if user, err = config.DBQueries.CreateUser(
			r.Context(),
			database.CreateUserParams{
				Email:          request.Email,
				HashedPassword: hash,
			},
		); err != nil {
			respJsonBadRequest(w, r, errorSomethingWentWrong)
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
			respJsonUnauthorized(w, r, errorMissingToken)
			return
		}
		if userId, err = auth.ValidateJWT(token, config.Secret); err != nil {
			respJsonUnauthorized(w, r, errorInvalidToken)
			return
		}
		request := struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}{}
		if json.NewDecoder(r.Body).Decode(&request) != nil {
			respJsonBadRequest(w, r, errorSomethingWentWrong)
			return
		}
		if hash, err = auth.HashPassword(request.Password); err != nil {
			respJsonBadRequest(w, r, errorSomethingWentWrong)
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
			respJsonBadRequest(w, r, errorSomethingWentWrong)
			return
		}
		respJsonUser(w, r, user, empty, empty)
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
			respJsonBadRequest(w, r, errorSomethingWentWrong)
			return
		}
		if user, err = config.DBQueries.GetUser(
			r.Context(),
			request.Email,
		); err != nil || auth.ValidateHash(request.Password, user.HashedPassword) != nil {
			respJsonUnauthorized(w, r, errorInvalidEmailPassword)
			return
		}
		if token, err = auth.MakeJWT(
			user.ID,
			config.Secret,
			time.Hour,
		); err != nil {
			respJsonBadRequest(w, r, errorSomethingWentWrong)
			return
		}
		if refreshToken, err = auth.MakeRefreshToken(); err != nil {
			respJsonBadRequest(w, r, errorSomethingWentWrong)
			return
		}
		if _, err = config.DBQueries.CreateRefreshToken(
			r.Context(),
			database.CreateRefreshTokenParams{
				Token:     refreshToken,
				UserID:    uuid.NullUUID{UUID: user.ID, Valid: true},
				ExpiresAt: time.Now().Add(time.Hour * hoursInDay * daysInMonth * 2),
			},
		); err != nil {
			respJsonBadRequest(w, r, errorSomethingWentWrong)
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
			respJsonUnauthorized(w, r, errorMissingRefreshToken)
			return
		}
		if user, err = config.DBQueries.GetUserFromRefreshToken(
			r.Context(),
			refreshToken,
		); err != nil || user.ID == uuid.Nil {
			respJsonUnauthorized(w, r, errorInvalidRefreshToken)
			return
		}
		if token, err = auth.MakeJWT(
			user.ID,
			config.Secret,
			time.Hour,
		); err != nil {
			respJsonBadRequest(w, r, errorSomethingWentWrong)
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
			respPlainBadRequest(w, r, errorMissingRefreshToken)
			return
		}
		if config.DBQueries.DeleteRefreshToken(
			r.Context(),
			refreshToken,
		) != nil {
			respPlainBadRequest(w, r, errorSomethingWentWrong)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func HandlerGetApiChirpsId(config *ApiConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error
		var chirpId uuid.UUID
		var chirp database.Chirp
		if chirpId, err = uuid.Parse(r.PathValue("id")); err != nil {
			respJsonBadRequest(w, r, errorSomethingWentWrong)
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
		var userId uuid.UUID
		var chirps []database.Chirp
		userIdParam := r.URL.Query().Get("author_id")
		if len(userIdParam) == 0 {
			if chirps, err = config.DBQueries.GetChirps(r.Context()); err != nil {
				respJsonBadRequest(w, r, errorSomethingWentWrong)
				return
			}
		} else {
			if userId, err = uuid.Parse(userIdParam); err != nil {
				respJsonBadRequest(w, r, errorSomethingWentWrong)
				return
			}
			if chirps, err = config.DBQueries.GetChirpsFromUser(
				r.Context(),
				uuid.NullUUID{UUID: userId, Valid: true},
			); err != nil {
				respJsonBadRequest(w, r, errorSomethingWentWrong)
				return
			}
		}
		if r.URL.Query().Get("sort") == orderDesc {
			slices.Reverse(chirps)
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
			respJsonUnauthorized(w, r, errorMissingToken)
			return
		}
		if userId, err = auth.ValidateJWT(token, config.Secret); err != nil {
			respJsonUnauthorized(w, r, errorInvalidToken)
			return
		}
		request := struct {
			Body string `json:"body"`
		}{}
		if json.NewDecoder(r.Body).Decode(&request) != nil {
			respJsonBadRequest(w, r, errorSomethingWentWrong)
			return
		}
		if len(request.Body) > 140 {
			respJsonBadRequest(w, r, errorChirpTooLong)
			return
		}
		if chirp, err = config.DBQueries.CreateChirp(
			r.Context(),
			database.CreateChirpParams{
				Body:   cleanProfanities(request.Body, profanities),
				UserID: uuid.NullUUID{UUID: userId, Valid: true},
			},
		); err != nil {
			respJsonBadRequest(w, r, errorSomethingWentWrong)
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
			respPlainBadRequest(w, r, errorSomethingWentWrong)
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
			respPlainBadRequest(w, r, errorSomethingWentWrong)
			return
		}
		w.WriteHeader(http.StatusNoContent)
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
			respPlainBadRequest(w, r, errorSomethingWentWrong)
			return
		}
		if request.Event != polkaEventUserUpgraded {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		if user, err = config.DBQueries.UpdateUserRed(
			r.Context(),
			request.Data.UserId,
		); err != nil || user.ID == uuid.Nil {
			respPlainNotFound(w, r)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}
