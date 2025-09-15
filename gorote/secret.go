package gorote

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

func GenerateJwtWithRSA(claims jwt.Claims, privateKey *rsa.PrivateKey) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	signedToken, err := token.SignedString(privateKey)
	if err != nil {
		return "", err
	}

	return signedToken, nil
}

func GenerateJwtWithSecret(claims jwt.Claims, secret string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}

	return signedToken, nil
}

func ValidateOrGetJWTRSA(claims jwt.Claims, hash string, publicKey *rsa.PublicKey) error {
	tokenString := strings.TrimPrefix(hash, "Bearer ")
	if tokenString == "" {
		return fmt.Errorf("authorization header is empty or malformed")
	}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return publicKey, nil
	})
	if err != nil || !token.Valid {
		return fmt.Errorf("invalid token")
	}
	return nil
}

func ValidateOrGetJWT(claims jwt.Claims, hash string, secret string) error {
	tokenString := strings.TrimPrefix(hash, "Bearer ")
	if tokenString == "" {
		return fmt.Errorf("authorization header is empty or malformed")
	}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil || !token.Valid {
		return fmt.Errorf("invalid token")
	}
	return nil
}

func GenerateFilePemInProject() error {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return err
	}
	privateKeyPEM := pem.EncodeToMemory(
		&pem.Block{
			Type:  "BEGIN PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
		},
	)
	publicKey := &privateKey.PublicKey
	publicKeyPEM := pem.EncodeToMemory(
		&pem.Block{
			Type:  "BEGIN PUBLIC KEY",
			Bytes: x509.MarshalPKCS1PublicKey(publicKey),
		},
	)
	if err = os.WriteFile("private_key.pem", privateKeyPEM, 0600); err != nil {
		return err
	}

	if err = os.WriteFile("public_key.pem", publicKeyPEM, 0600); err != nil {
		return err
	}
	return nil
}

func MustReadPrivateKeyFromFile(filePath string) *rsa.PrivateKey {
	pemBytes, err := os.ReadFile(filePath)
	if err != nil {
		panic("failed to read public key")
	}

	block, _ := pem.Decode(pemBytes)
	if block == nil {
		panic("failed to decode PEM block")
	}

	privateKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		privateKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			panic("failed to parse private key")
		}
	}

	rsaPriv, ok := privateKey.(*rsa.PrivateKey)
	if !ok {
		panic("not an RSA private key")
	}

	return rsaPriv
}

func MustReadPublicKeyFromFile(filePath string) *rsa.PublicKey {
	pemBytes, err := os.ReadFile(filePath)
	if err != nil {
		panic("failed to read public key")
	}

	block, _ := pem.Decode(pemBytes)
	if block == nil {
		panic("failed to decode PEM block")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		pub, err = x509.ParsePKCS1PublicKey(block.Bytes)
		if err != nil {
			panic("failed to parse public key")
		}
	}

	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		panic("not an RSA public key")
	}

	return rsaPub
}

func MustReadPublicKeyFromString(key string) *rsa.PublicKey {
	derBytes, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		panic("failed to decode public key")
	}

	pub, err := x509.ParsePKIXPublicKey(derBytes)
	if err != nil {
		pub, err = x509.ParsePKCS1PublicKey(derBytes)
		if err != nil {
			panic("failed to parse public key")
		}
	}

	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		panic("not an RSA public key")
	}

	return rsaPub
}

func MustReadPrivateKeyFromString(key string) *rsa.PrivateKey {
	derBytes, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		panic("failed to decode private key")
	}

	privateKey, err := x509.ParsePKCS8PrivateKey(derBytes)
	if err != nil {
		privateKey, err = x509.ParsePKCS1PrivateKey(derBytes)
		if err != nil {
			panic("failed to parse private key")
		}
	}

	rsaPriv, ok := privateKey.(*rsa.PrivateKey)
	if !ok {
		panic("not an RSA private key")
	}

	return rsaPriv
}
