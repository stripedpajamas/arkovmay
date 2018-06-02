package controllers

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/oklog/ulid"
	"github.com/stripedpajamas/arkovmay/database"
	"github.com/stripedpajamas/arkovmay/database/models"
)

type CreateRequest struct {
	Name string `json"name"`
}

type MarkResponse struct {
	ID       uint   `json:"id"`
	Name     string `json:"name"`
	PublicID string `json:"publicId"`
}

func MarkCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		markID := chi.URLParam(r, "markID")
		var mark models.Mark
		database.DB.Where("id = ?", markID).First(&mark)
		if mark.ID == 0 {
			http.Error(w, http.StatusText(404), 404)
			return
		}
		ctx := context.WithValue(r.Context(), "mark", mark)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func Generate(w http.ResponseWriter, r *http.Request) {

}

func CreateMark(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user, ok := ctx.Value("user").(models.User)
	if !ok {
		http.Error(w, http.StatusText(422), 422)
		return
	}

	// want to create a new mark for this user
	// need to parse the body first
	decoder := json.NewDecoder(r.Body)
	var cr CreateRequest
	err := decoder.Decode(&cr)
	if err != nil || cr.Name == "" {
		http.Error(w, http.StatusText(400), 400)
		return
	}
	defer r.Body.Close()

	// find or init a new mark record for user
	var mark models.Mark
	database.DB.Where(models.Mark{Name: cr.Name, UserID: user.ID}).FirstOrInit(&mark)

	// not already found, so populate fields
	if mark.PublicID == "" {
		// generate a ulid
		u, err := ulid.New(ulid.Timestamp(time.Now()), rand.Reader)
		if err != nil {
			http.Error(w, http.StatusText(500), 500)
			return
		}
		mark.PublicID = u.String()

		// create record
		database.DB.Create(&mark)

		createResponse, err := json.Marshal(MarkResponse{
			Name:     mark.Name,
			PublicID: mark.PublicID,
		})
		if err != nil {
			http.Error(w, http.StatusText(500), 500)
			return
		}

		w.WriteHeader(http.StatusCreated)
		w.Write(createResponse)
		return
	}

	// record already existed
	w.WriteHeader(http.StatusConflict)
}

func GetAllMarks(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user, ok := ctx.Value("user").(models.User)
	if !ok {
		http.Error(w, http.StatusText(422), 422)
		return
	}

	// get just names and public ids for this user
	var marks []MarkResponse
	database.DB.Table("marks").Select("id, name, public_id").Where("user_id = ?", user.ID).Scan(&marks)

	response, err := json.Marshal(&marks)
	if err != nil {
		http.Error(w, http.StatusText(500), 500)
		return
	}
	w.Write(response)
}

func GetMark(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	mark, ok := ctx.Value("mark").(models.Mark)
	if !ok {
		http.Error(w, http.StatusText(422), 422)
		return
	}

	response, err := json.Marshal(&MarkResponse{
		ID:       mark.ID,
		Name:     mark.Name,
		PublicID: mark.PublicID,
	})
	if err != nil {
		http.Error(w, http.StatusText(500), 500)
		return
	}
	w.Write(response)
}

func UpdateMark(w http.ResponseWriter, r *http.Request) {

}

func DeleteMark(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	mark, ok := ctx.Value("mark").(models.Mark)
	if !ok {
		http.Error(w, http.StatusText(422), 422)
		return
	}

	database.DB.Delete(&mark)
	w.WriteHeader(http.StatusOK)
}
