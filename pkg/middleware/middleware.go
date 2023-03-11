package middleware

import (
	"context"
	"net/http"

	"go.mongodb.org/mongo-driver/mongo"
)

// Middleware function to pass the MongoDB client variable to the routes
func MongoMiddleware(client *mongo.Client) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), "mongo", client)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
