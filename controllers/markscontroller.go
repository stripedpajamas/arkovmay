package controllers

import (
	"context"
	"net/http"
)

func MarkCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// markID := chi.URLParam(r, "markID")
		// TODO get mark from database
		// mark, err := dbGetArticle(markID)
		// if err != nil {
		// 	http.Error(w, http.StatusText(404), 404)
		// 	return
		// }
		mark := "help"
		ctx := context.WithValue(r.Context(), "mark", mark)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func Generate(w http.ResponseWriter, r *http.Request) {

}

func CreateMark(w http.ResponseWriter, r *http.Request) {

}

func GetAllMarks(w http.ResponseWriter, r *http.Request) {
	// ctx := r.Context()
	// user, ok := ctx.Value("user").(models.User)
	// if !ok {
	// 	http.Error(w, http.StatusText(422), 422)
	// 	return
	// }
	// w.WriteHeader(200)
}

func GetMark(w http.ResponseWriter, r *http.Request) {

}

func UpdateMark(w http.ResponseWriter, r *http.Request) {

}

func DeleteMark(w http.ResponseWriter, r *http.Request) {

}
