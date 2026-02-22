package authentication

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type IJwtService interface {
	GenerateToken(claims JWTClaims) (string, error)
	ParseToken(token string) (*JWTClaims, error)
	RefreshToken(token string) (string, error)
	ValidateToken(token string) error
}

type JWTClaims map[string]interface{}
type JWTConfig struct {
	SecretKey      string
	ExpirationTime int
	RefreshTime    int
	Issuer         string
	Audience       string
}

type jwtAuth struct {
	secretKey      []byte
	issuer         string
	audience       string
	refreshTime    int
	expirationTime int
}

func NewJWTService(opts *JWTConfig) IJwtService {
	return &jwtAuth{
		secretKey:      []byte(opts.SecretKey),
		expirationTime: opts.ExpirationTime,
		refreshTime:    opts.RefreshTime,
		issuer:         opts.Issuer,
		audience:       opts.Audience,
	}
}

func (j *jwtAuth) GenerateToken(dataClaims JWTClaims) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)

	now := time.Now()
	exp := now.Add(time.Duration(j.expirationTime) * time.Minute).Unix()

	claimsMap := jwt.MapClaims{
		"iss": j.issuer,
		"aud": j.audience,
		"exp": exp,
	}

	for key, value := range dataClaims {
		claimsMap[key] = value
	}

	token.Claims = claimsMap

	tokenString, err := token.SignedString(j.secretKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (j *jwtAuth) ParseToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return j.secretKey, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		jwtClaims := JWTClaims{}
		for key, value := range claims {
			jwtClaims[key] = value
		}
		return &jwtClaims, nil
	}

	return nil, errors.New("invalid token")
}

func (j *jwtAuth) RefreshToken(tokenString string) (string, error) {
	claims, err := j.ParseToken(tokenString)
	if err != nil {
		return "", err
	}

	newClaims := jwt.MapClaims{
		"iss": j.issuer,
		"aud": j.audience,
		"exp": time.Now().Add(time.Duration(j.refreshTime) * time.Minute).Unix(),
	}

	for key, value := range *claims {
		if key != "exp" && key != "iss" && key != "aud" {
			newClaims[key] = value
		}
	}

	newToken := jwt.NewWithClaims(jwt.SigningMethodHS256, newClaims)

	newTokenString, err := newToken.SignedString(j.secretKey)
	if err != nil {
		return "", err
	}

	return newTokenString, nil
}

func (j *jwtAuth) ValidateToken(tokenString string) error {
	_, err := j.ParseToken(tokenString)
	return err
}
