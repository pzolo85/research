package main

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func main() {

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodES512, jwt.RegisteredClaims{
		Issuer:    "issuer",
		Subject:   "subject",
		Audience:  []string{"aud1", "aud2"},
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24)),
		NotBefore: jwt.NewNumericDate(time.Now()),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ID:        uuid.NewString(),
	})

	// openssl ecparam -name secp521r1 -genkey -noout -out ec-p512-private.pem
	// openssl ec -in ec-p512-private.pem -pubout -out ec-p512-public.pem

	pri, err := os.ReadFile("./ec-p512-private.pem")
	if err != nil {
		panic(err)
	}

	blo, rest := pem.Decode(pri)
	if len(rest) != 0 {
		panic(rest)
	}

	ecPrivKey, err := x509.ParseECPrivateKey(blo.Bytes)
	if err != nil {
		panic(err)
	}

	token, err := jwtToken.SignedString(ecPrivKey)
	if err != nil {
		panic(err)
	}

	fmt.Println(token)

}
