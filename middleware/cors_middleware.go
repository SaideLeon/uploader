package middleware

import (
	"net/http"
)

// CORSMiddleware adiciona os cabeçalhos CORS necessários para permitir requisições de diferentes origens.
func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Define o cabeçalho 'Access-Control-Allow-Origin' para permitir qualquer origem.
		// Em um ambiente de produção, você pode querer restringir isso a domínios específicos.
		w.Header().Set("Access-Control-Allow-Origin", "*")

		// Define os métodos HTTP permitidos.
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")

		// Define os cabeçalhos permitidos.
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		// Se a requisição for um 'OPTIONS' (preflight request), apenas retorne os cabeçalhos.
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Continua para o próximo handler na cadeia.
		next.ServeHTTP(w, r)
	})
}
