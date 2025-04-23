package core

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type PayloadJwt struct {
	Token  string
	Claims JwtClaims
}
type JwtClaims struct {
	Sub         uint     `json:"sub"`
	Exp         int      `json:"exp"`
	Permissions []string `json:"permissions"`
	IsSuperUser bool     `json:"isSuperUser"`
	jwt.RegisteredClaims
}

type GenToken struct {
	Id          uint
	AppName     string
	Permissions []string
	IsSuperUser bool
	TimeZone    string
	PrivateKey  *rsa.PrivateKey
	Ttl         time.Duration
}

func GenerateTokenRSA(gen *GenToken) (string, error) {
	location, err := time.LoadLocation(gen.TimeZone)
	if err != nil {
		return "", fmt.Errorf("invalid timezone: %s", err.Error())
	}
	currentTime := time.Now().In(location)

	accessTokenExpirationTime := currentTime.Add(gen.Ttl)

	accessClaims := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"sub":         gen.Id,
		"iss":         gen.AppName,
		"permissions": gen.Permissions,
		"isSuperUser": gen.IsSuperUser,
		"iat":         currentTime.Unix(),
		"exp":         accessTokenExpirationTime.Unix(),
	})

	accessToken, err := accessClaims.SignedString(gen.PrivateKey)
	if err != nil {
		return "", fmt.Errorf("could not sign access token string %v", err.Error())
	}

	return accessToken, nil
}

func GetJwtHeaderPayloadRSA(authHeader string, publicKey *rsa.PublicKey) (*PayloadJwt, error) {
	// Extrair o token da string de autorização
	tokenString := strings.TrimPrefix(authHeader, "Bearer ")

	// Validar o token usando a chave pública
	token, err := jwt.ParseWithClaims(tokenString, &JwtClaims{}, func(token *jwt.Token) (any, error) {
		// Verificar o método de assinatura
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return publicKey, nil
	})
	if err != nil {
		return nil, err
	}

	// Verificar se o token é válido
	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	// Obter as reivindicações do token
	claims, ok := token.Claims.(*JwtClaims)
	if !ok {
		return nil, fmt.Errorf("invalid claims type")
	}

	jwt := &PayloadJwt{
		Token:  tokenString,
		Claims: *claims,
	}

	return jwt, nil
}

func GeneratePemRSA() error {
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

func ReadRSAPrivateKeyFromFile(filePath string) (*rsa.PrivateKey, error) {
	pemBytes, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, errors.New("failed to decode PEM block")
	}

	// Tenta parsear como PKCS#8 (formato padrão do OpenSSL genpkey)
	privateKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		// Se falhar, tenta parsear como PKCS#1 (formato tradicional)
		privateKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse private key: %v", err)
		}
	}

	rsaPriv, ok := privateKey.(*rsa.PrivateKey)
	if !ok {
		return nil, errors.New("not an RSA private key")
	}

	return rsaPriv, nil
}

func ReadRSAPublicKeyFromFile(filePath string) (*rsa.PublicKey, error) {
	pemBytes, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, errors.New("failed to decode PEM block")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		// Se falhar, tenta parsear como PKCS#1 (algumas chaves públicas podem estar nesse formato)
		rsaPub, err := x509.ParsePKCS1PublicKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse public key: %v", err)
		}
		return rsaPub, nil
	}

	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("not an RSA public key")
	}

	return rsaPub, nil
}
