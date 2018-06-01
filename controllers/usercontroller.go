package controllers

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"net/http"
	"time"

	"github.com/oklog/ulid"
	"github.com/stripedpajamas/arkovmay/database"
	"github.com/stripedpajamas/arkovmay/database/models"
)

const sessionCookie = "ARKOVMAY_SESSION"

// minutes
var loginExpiryMinutes = time.Minute * 1

type TokenRequest struct {
	Email string `json:"email"`
}

type LoginRequest struct {
	Email string `json:"email"`
	Token string `json:"token"`
}

type LoginResponse struct {
	Email string `json:"email"`
	Token string `json:"token"`
}

func CreateUser(w http.ResponseWriter, r *http.Request) {
	// registration is the same as requesting a login
	// TODO rate limit so they can only call this 1/5m

	// get email from body
	decoder := json.NewDecoder(r.Body)
	var tr TokenRequest
	err := decoder.Decode(&tr)
	if err != nil || tr.Email == "" {
		http.Error(w, http.StatusText(400), 400)
		return
	}
	defer r.Body.Close()

	// generate a ulid
	u, err := ulid.New(ulid.Timestamp(time.Now()), rand.Reader)
	if err != nil {
		http.Error(w, http.StatusText(500), 500)
		return
	}

	// add login token to db
	tokenModel := &models.LoginToken{
		Email:   tr.Email,
		Token:   u.String(),
		Expires: time.Now().Add(loginExpiryMinutes),
	}
	database.DB.Create(tokenModel)

	// TODO send email

	w.WriteHeader(http.StatusOK)
}

func Login(w http.ResponseWriter, r *http.Request) {
	// get email from body
	decoder := json.NewDecoder(r.Body)
	var lr LoginRequest
	err := decoder.Decode(&lr)
	if err != nil || lr.Email == "" || lr.Token == "" {
		http.Error(w, http.StatusText(400), 400)
		return
	}
	defer r.Body.Close()

	// see if token is valid and hasn't expired
	var foundToken models.LoginToken
	database.DB.Where("email = ? AND token = ?", lr.Email, lr.Token).First(&foundToken)

	// token not found
	if foundToken.ID == 0 {
		http.Error(w, http.StatusText(401), 401)
		return
	}

	// token has expired :(
	if time.Until(foundToken.Expires) < 0 {
		// delete that token
		database.DB.Unscoped().Delete(&foundToken)
		http.Error(w, http.StatusText(401), 401)
		return
	}

	// delete token
	database.DB.Unscoped().Delete(&foundToken)

	// all good -- see if the email already exists in users table
	// otherwise create it
	var user models.User
	database.DB.Where(models.User{Email: lr.Email}).FirstOrCreate(&user)

	// create session token
	u, err := ulid.New(ulid.Timestamp(time.Now()), rand.Reader)
	if err != nil {
		http.Error(w, http.StatusText(500), 500)
		return
	}

	// add session token to db
	tokenModel := &models.SessionToken{
		UserID: user.ID,
		Token:  u.String(),
	}
	database.DB.Create(tokenModel)

	// send back token to client
	loginResponse, err := json.Marshal(LoginResponse{
		Email: user.Email,
		Token: tokenModel.Token,
	})
	if err != nil {
		http.Error(w, http.StatusText(500), 500)
		return
	}

	w.Write(loginResponse)
}

func Logout(w http.ResponseWriter, r *http.Request) {
	// see if the session token is in a cookie
	s, _ := r.Cookie(sessionCookie)
	// we can ignore errors (not found), we're 200 either way
	var session models.SessionToken
	database.DB.Where("token = ?", s.Value).First(&session)
	if session.ID != 0 {
		database.DB.Unscoped().Delete(&session)
	}
	w.WriteHeader(200)
}

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// confirm that the session cookie that came in is valid
		// that is-- in the db, and that has not expired
		s, err := r.Cookie(sessionCookie)
		if err != nil {
			http.Error(w, http.StatusText(401), 401)
			return
		}
		var session models.SessionToken
		database.DB.Where("token = ?", s.Value).First(&session)

		// session not found
		if session.ID == 0 {
			http.Error(w, http.StatusText(401), 401)
			return
		}

		// just confirm the user exists and attach it to the session
		var user models.User
		database.DB.Where("ID = ?", session.UserID).First(&user)

		// user not found
		if user.ID == 0 {
			http.Error(w, http.StatusText(401), 401)
			return
		}

		// attach to context
		ctx := context.WithValue(r.Context(), "user", user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
