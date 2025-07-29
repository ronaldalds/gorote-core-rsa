package gorote

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

func SetTokenCookie(ctx *fiber.Ctx, domains, typeToken, token string, ttl time.Duration) error {
	if domains != "" {
		for domain := range strings.SplitSeq(domains, ",") {
			domain = strings.TrimSpace(domain)
			if domain == "" {
				continue
			}
			ctx.Cookie(&fiber.Cookie{
				Name:     typeToken,
				Value:    token,
				HTTPOnly: true,
				Secure:   true,
				SameSite: "None",
				Domain:   domain,
				Path:     "/",
				MaxAge:   int(ttl.Seconds()),
			})
		}
	}
	return nil
}

func GenerateTokenRSA(claims jwt.Claims, privateKey *rsa.PrivateKey) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	signedToken, err := token.SignedString(privateKey)
	if err != nil {
		return "", fmt.Errorf("could not sign access token: %v", err)
	}

	return signedToken, nil
}

func GenerateToken(claims jwt.Claims, secret string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", fmt.Errorf("could not sign access token: %v", err)
	}

	return signedToken, nil
}

func ValidateOrGetJWTRSA(claims jwt.Claims, authHeader string, publicKey *rsa.PublicKey) error {
	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
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

func ValidateOrGetJWT(claims jwt.Claims, authHeader string, secret string) error {
	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
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
