package middleware

import (
	"net/http"
	"strings"

	"github.com/go-chi/cors"
)

func CORSMiddleware(allowedOrigins []string) cors.Options {
	exact := make(map[string]struct{}, len(allowedOrigins))
	var suffixes []string
	for _, o := range allowedOrigins {
		if strings.Contains(o, "*") {
			s := strings.TrimPrefix(o, "https://*")
			s = strings.TrimPrefix(s, "http://*")
			suffixes = append(suffixes, s)
		} else {
			exact[o] = struct{}{}
		}
	}

	return cors.Options{
		AllowOriginFunc: func(_ *http.Request, origin string) bool {
			if _, ok := exact[origin]; ok {
				return true
			}
			for _, s := range suffixes {
				if strings.HasSuffix(origin, s) {
					return true
				}
			}
			return false
		},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}
}
