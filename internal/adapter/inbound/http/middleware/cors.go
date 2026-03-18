package middleware

import (
	"os"
	"strings"

	"github.com/go-chi/cors"
)

func CorsHandler() cors.Options {
	origins := []string{"http://localhost:5173", "http://localhost:3000"}
	if extra := os.Getenv("ALLOWED_ORIGINS"); extra != "" {
		origins = append(origins, strings.Split(extra, ",")...)
	}
	return cors.Options{
		AllowedOrigins:   origins,
		AllowedMethods:   []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
		MaxAge:           300,
	}
}
