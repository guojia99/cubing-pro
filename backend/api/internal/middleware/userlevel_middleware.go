package middleware

import "net/http"

type UserLevelMiddleware struct {
}

func NewUserLevelMiddleware() *UserLevelMiddleware {
	return &UserLevelMiddleware{}
}

func (m *UserLevelMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO generate middleware implement function, delete after code implementation

		// Passthrough to next handler if need
		next(w, r)
	}
}
