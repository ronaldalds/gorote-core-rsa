package core

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

type JwtClaims struct {
	Sub         uint     `json:"sub"`
	Email       string   `json:"email"`
	IsSuperUser bool     `json:"isSuperUser"`
	Permissions []string `json:"permissions"`
	Tenants     []uint   `json:"tenants"`
	Type        string   `json:"type"`
	jwt.RegisteredClaims
}

type GenToken struct {
	Id          uint
	Email       string
	AppName     string
	Permissions []string
	Tenants     []uint
	IsSuperUser bool
	TimeZone    string
	Type        string
	PrivateKey  *rsa.PrivateKey
	Ttl         time.Duration
}

type ConfigToken struct {
	TokenType   string
	AppName     string
	AppTimeZone string
	Domain      string
	PrivateKey  *rsa.PrivateKey
	Ttl         time.Duration
}

func SetToken(ctx *fiber.Ctx, user *User, config *ConfigToken) (string, error) {
	permissions := ExtractCodePermissionsByUser(user)

	var tenants []uint
	for _, tenant := range user.Tenants {
		tenants = append(tenants, tenant.ID)
	}
	accessToken, err := GenerateTokenRSA(&GenToken{
		Id:          user.ID,
		Email:       user.Email,
		AppName:     config.AppName,
		Permissions: permissions,
		IsSuperUser: user.IsSuperUser,
		Tenants:     tenants,
		TimeZone:    config.AppTimeZone,
		PrivateKey:  config.PrivateKey,
		Ttl:         config.Ttl,
		Type:        config.TokenType,
	})
	if err != nil {
		return "", err
	}
	if config.Domain != "" {
		domains := strings.Split(config.Domain, ",")
		for _, domain := range domains {
			domain = strings.TrimSpace(domain)
			if domain == "" {
				continue
			}

			ctx.Cookie(&fiber.Cookie{
				Name:     config.TokenType,
				Value:    accessToken,
				HTTPOnly: true,
				Secure:   true,
				SameSite: "None",
				Domain:   domain,
				Path:     "/",
				MaxAge:   int(config.Ttl.Seconds()),
			})
		}
	}
	return accessToken, nil
}

func GenerateTokenRSA(gen *GenToken) (string, error) {
	location, err := time.LoadLocation(gen.TimeZone)
	if err != nil {
		return "", fmt.Errorf("invalid timezone: %s", err.Error())
	}
	currentTime := time.Now().In(location)
	expirationTime := currentTime.Add(gen.Ttl)

	claims := JwtClaims{
		Sub:         gen.Id,
		Email:       gen.Email,
		Permissions: gen.Permissions,
		IsSuperUser: gen.IsSuperUser,
		Tenants:     gen.Tenants,
		Type:        gen.Type,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    gen.AppName,
			IssuedAt:  jwt.NewNumericDate(currentTime),
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	signedToken, err := token.SignedString(gen.PrivateKey)
	if err != nil {
		return "", fmt.Errorf("could not sign access token: %v", err)
	}

	return signedToken, nil
}

func GetJwtHeaderPayloadRSA(authHeader string, publicKey *rsa.PublicKey) (*JwtClaims, error) {
	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	if tokenString == "" {
		return nil, fmt.Errorf("authorization header is empty or malformed")
	}

	token, err := jwt.ParseWithClaims(tokenString, &JwtClaims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return publicKey, nil
	})
	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(*JwtClaims)
	if !ok {
		return nil, fmt.Errorf("invalid claims type")
	}

	return claims, nil
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

	privateKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
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

func ReadRSAPublicKeyFromString(key string) (*rsa.PublicKey, error) {
	derBytes, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		return nil, err
	}

	pub, err := x509.ParsePKIXPublicKey(derBytes)
	if err != nil {
		rsaPub, err := x509.ParsePKCS1PublicKey(derBytes)
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

func ReadRSAPrivateKeyFromString(key string) (*rsa.PrivateKey, error) {
	derBytes, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		return nil, err
	}

	privateKey, err := x509.ParsePKCS8PrivateKey(derBytes)
	if err != nil {
		privateKey, err = x509.ParsePKCS1PrivateKey(derBytes)
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
