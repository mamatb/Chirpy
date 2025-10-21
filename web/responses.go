package web

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/mamatb/Chirpy/database"
)

func respPlainOk(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set(headerContentType, contentTypePlain)
	body := []byte(httpOkPlain)
	if _, err := w.Write(body); err != nil {
		log.Fatal(err)
	}
}

func respPlainBadRequest(w http.ResponseWriter, _ *http.Request, message string) {
	w.WriteHeader(http.StatusBadRequest)
	w.Header().Set(headerContentType, contentTypePlain)
	body := []byte(message)
	if _, err := w.Write(body); err != nil {
		log.Fatal(err)
	}
}

func respPlainUnauthorized(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusUnauthorized)
	w.Header().Set(headerContentType, contentTypePlain)
	body := []byte(httpUnauthorizedPlain)
	if _, err := w.Write(body); err != nil {
		log.Fatal(err)
	}
}

func respPlainForbidden(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusForbidden)
	w.Header().Set(headerContentType, contentTypePlain)
	body := []byte(httpForbiddenPlain)
	if _, err := w.Write(body); err != nil {
		log.Fatal(err)
	}
}

func respPlainNotFound(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	w.Header().Set(headerContentType, contentTypePlain)
	body := []byte(httpNotFoundPlain)
	if _, err := w.Write(body); err != nil {
		log.Fatal(err)
	}
}

func respJsonUser(w http.ResponseWriter, _ *http.Request, user database.User,
	token string, refreshToken string) {
	w.Header().Set(headerContentType, contentTypeJson)
	var err error
	var body []byte
	if body, err = json.Marshal(jsonUser{
		Id:           user.ID,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
		Email:        user.Email,
		IsChirpyRed:  user.IsChirpyRed,
		Token:        token,
		RefreshToken: refreshToken,
	}); err != nil {
		log.Fatal(err)
	}
	if _, err = w.Write(body); err != nil {
		log.Fatal(err)
	}
}

func respJsonToken(w http.ResponseWriter, _ *http.Request, token string) {
	w.Header().Set(headerContentType, contentTypeJson)
	var err error
	var body []byte
	if body, err = json.Marshal(jsonToken{
		Token: token,
	}); err != nil {
		log.Fatal(err)
	}
	if _, err = w.Write(body); err != nil {
		log.Fatal(err)
	}
}

func respJsonChirp(w http.ResponseWriter, _ *http.Request, chirp database.Chirp) {
	w.Header().Set(headerContentType, contentTypeJson)
	var err error
	var body []byte
	if body, err = json.Marshal(jsonChirp{
		Id:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserId:    chirp.UserID,
	}); err != nil {
		log.Fatal(err)
	}
	if _, err = w.Write(body); err != nil {
		log.Fatal(err)
	}
}

func respJsonChirps(w http.ResponseWriter, _ *http.Request, chirps []database.Chirp) {
	w.Header().Set(headerContentType, contentTypeJson)
	var err error
	var body []byte
	var chirpsJson []jsonChirp
	for _, chirp := range chirps {
		chirpsJson = append(chirpsJson, jsonChirp{
			Id:        chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body:      chirp.Body,
			UserId:    chirp.UserID,
		})
	}
	if body, err = json.Marshal(chirpsJson); err != nil {
		log.Fatal(err)
	}
	if _, err = w.Write(body); err != nil {
		log.Fatal(err)
	}
}

func respJsonUserCreated(w http.ResponseWriter, r *http.Request, user database.User) {
	w.WriteHeader(http.StatusCreated)
	respJsonUser(w, r, user, empty, empty)
}

func respJsonChirpCreated(w http.ResponseWriter, r *http.Request, chirp database.Chirp) {
	w.WriteHeader(http.StatusCreated)
	respJsonChirp(w, r, chirp)
}

func respJsonBadRequest(w http.ResponseWriter, _ *http.Request, message string) {
	w.WriteHeader(http.StatusBadRequest)
	w.Header().Set(headerContentType, contentTypeJson)
	var err error
	var body []byte
	if body, err = json.Marshal(jsonError{
		Error: message,
	}); err != nil {
		log.Fatal(err)
	}
	if _, err = w.Write(body); err != nil {
		log.Fatal(err)
	}
}

func respJsonUnauthorized(w http.ResponseWriter, _ *http.Request, message string) {
	w.WriteHeader(http.StatusUnauthorized)
	w.Header().Set(headerContentType, contentTypeJson)
	var err error
	var body []byte
	if body, err = json.Marshal(jsonError{
		Error: message,
	}); err != nil {
		log.Fatal(err)
	}
	if _, err = w.Write(body); err != nil {
		log.Fatal(err)
	}
}
