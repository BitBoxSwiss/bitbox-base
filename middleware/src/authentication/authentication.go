package authentication

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"log"
	"time"

	"github.com/dgrijalva/jwt-go"
)

// Claims is a struct that is consumed by the jwt-go that includes the claims that need to be included into a json web token.
type Claims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

// JwtAuth is a struct holding the jwtKey. The key is generated once when starting the middleware.
type JwtAuth struct {
	jwtKey string
}

// NewJwtAuth returns a new JwtAuth struct if the jwt key generation is successful, or an error, if it is not.
func NewJwtAuth() (*JwtAuth, error) {
	jwtAuth := &JwtAuth{}
	// generate random string with 32 bytes of entropy and use it as the signing key
	JwtKey, err := jwtAuth.generateRandomString(32)
	if err != nil {
		log.Println("could not get enough entropy to generate key")
		return &JwtAuth{}, err
	}
	jwtAuth.jwtKey = JwtKey
	return jwtAuth, nil
}

// generateRandomBytes returns securely generated random bytes. It will return an error
// if the system's secure random number generator fails to function correctly, in which
// case the caller should not continue. These utility functions are taken from:
// https://stackoverflow.com/questions/32349807/how-can-i-generate-a-random-int-using-the-crypto-rand-package
func (jwtAuth *JwtAuth) generateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	// Note that err == nil only if we read len(b) bytes.
	if err != nil {
		log.Println("unable to read random bytes during jwt key generation")
		return nil, err
	}

	return b, nil
}

// generateRandomString returns a URL-safe, base64 encoded securely generated random string.
func (jwtAuth *JwtAuth) generateRandomString(s int) (string, error) {
	b, err := jwtAuth.generateRandomBytes(s)
	return base64.URLEncoding.EncodeToString(b), err
}

// GenerateToken generates a jwt token. It is the callers job to ensure that the username has been verified in the database first.
func (jwtAuth *JwtAuth) GenerateToken(username string) (string, error) {
	jwtStaticKey := []byte(jwtAuth.jwtKey)

	expirationTime := time.Now().Add(24 * time.Hour)

	claims := &Claims{
		Username: username,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtStaticKey)
	if err != nil {
		log.Println("error generating new tokenString from middleware jwt static session key")
		return "", err
	}
	return tokenString, nil
}

// ValidateToken takes a jwt token string as an argument and returns an error if the validation fails. If the validation is successful, nil is returned
func (jwtAuth *JwtAuth) ValidateToken(tokenStr string) error {
	jwtStaticKey := []byte(jwtAuth.jwtKey)

	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtStaticKey, nil
	})
	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			log.Println("Invalid Signature")
		}
		return err
	}
	if !token.Valid {
		log.Println("Invalid token received, breaking connection with: ", token.Claims)
		return errors.New("invalid jwt token received")
	}
	return nil
}
