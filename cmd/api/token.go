package main


import (
	"time"
	"github.com/AbrahamMayowa/ticketmania/internal/data"
	jwt "github.com/golang-jwt/jwt/v5"
	"fmt"
)

type TokenScope string

const (
	ScopeAuthentication TokenScope = "authentication"
)

type AuthClaims struct {
	Scope TokenScope `json:"scope"`
	Email string     `json:"email"`
	Id	int64  `json:"id"`
	jwt.RegisteredClaims
}

func (app *application) GenerateToken(scope TokenScope, user *data.User) (string, error) {



claims := AuthClaims{
	Scope: scope,
	Email: user.Email,
	Id:    user.Id,
	RegisteredClaims: jwt.RegisteredClaims{
		Issuer:    "github.com/AbrahamMayowa/ticketmania",
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		NotBefore: jwt.NewNumericDate(time.Now()),
	},
}

token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
ss, err := token.SignedString([]byte(app.config.jwt.secret))

if err != nil {
	return "", err
}
return string(ss), nil
}


func (app *application) ValidateToken(tokenString string) (*AuthClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &AuthClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(app.config.jwt.secret), nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*AuthClaims)
	if !ok || !token.Valid {	
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}

