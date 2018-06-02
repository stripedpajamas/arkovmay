package controllers

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/stripedpajamas/arkovmay/builder"

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
			ID:       mark.ID,
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
	database.DB.Table("marks").Select("id, name, public_id").Where("user_id = ? AND deleted_at IS NULL", user.ID).Scan(&marks)

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
	// should use builder to update data of a mark with posted text file
	ctx := r.Context()
	mark, ok := ctx.Value("mark").(models.Mark)
	if !ok {
		http.Error(w, http.StatusText(422), 422)
		return
	}

	if err := r.ParseMultipartForm(32 << 20); nil != err {
		http.Error(w, http.StatusText(500), 500)
		return
	}
	var filedata [][]byte
	for _, fileheaders := range r.MultipartForm.File {
		for _, header := range fileheaders {
			file, err := header.Open()
			if err != nil {
				http.Error(w, http.StatusText(500), 500)
				return
			}
			defer file.Close()
			var data []byte
			if data, err = ioutil.ReadAll(file); err != nil {
				http.Error(w, http.StatusText(500), 500)
				return
			}
			filedata = append(filedata, data)
		}
	}

	// we have the bytes in filedata now. we need to get the current map from the db
	var wordMap map[string]map[string]*builder.Node
	oldData := mark.Data
	// handle no old data
	if oldData == "" {
		oldData = "{}"
	}
	if err := json.Unmarshal([]byte(oldData), &wordMap); err != nil {
		fmt.Println(err)
		http.Error(w, http.StatusText(500), 500)
		return
	}

	input := string(bytes.Join(filedata, []byte(" ")))
	wordMap = builder.Build(string(input), wordMap)

	encoded, err := json.Marshal(wordMap)
	if err != nil {
		http.Error(w, http.StatusText(500), 500)
		return
	}

	// write wordmap to database
	database.DB.Model(&mark).Update("data", encoded)
	w.Write(encoded)
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
