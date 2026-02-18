package middleware

import (
	"net/http"

	appconfig "admin.com/admin-api/config"
)

func CORSMiddleware(next http.Handler, corsCfg appconfig.CORSConfig) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", corsCfg.AllowOrigin)
		w.Header().Set("Access-Control-Allow-Methods", corsCfg.AllowMethods)
		w.Header().Set("Access-Control-Allow-Headers", corsCfg.AllowHeaders)

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
