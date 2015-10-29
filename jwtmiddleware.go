package main

import (
	"github.com/auth0/go-jwt-middleware"
	"github.com/dgrijalva/jwt-go"
	"net/http"
)

func secretHandler(token *jwt.Token) (interface{}, error) {
	//TODO auth-server nach key fragen.
	return []byte("My Secret"), nil
}

func JwtMiddleware(next http.Handler) http.Handler {
	jwtMiddleware := jwtmiddleware.New(jwtmiddleware.Options{
		ValidationKeyGetter: secretHandler,
	})
	return jwtMiddleware.Handler(next)
}
