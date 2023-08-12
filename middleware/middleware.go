package middleware

import (
	"net/http"

	"github.com/KasperKornak/StockApp/sessions"
)

// require auth
func AuthRequired(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// if session doesnt exist redirect to /login
		session, _ := sessions.Store.Get(r, "session")
		_, ok := session.Values["user_id"]
		if !ok {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		handler.ServeHTTP(w, r)
	}
}
