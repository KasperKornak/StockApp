package middleware

import (
	"net/http"

	"github.com/KasperKornak/StockApp/sessions"
)

func AuthRequired(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, _ := sessions.Store.Get(r, "session")
		_, ok := session.Values["user_id"]
		if !ok {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		handler.ServeHTTP(w, r)
	}
}
