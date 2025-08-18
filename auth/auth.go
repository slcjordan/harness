package auth

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/slcjordan/harness"
)

type JWT struct {
	Secret []byte
}

func (j JWT) Handle(ctx context.Context, camera harness.Camera) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"aud":  "video-analytics-193323",
		"exp":  time.Now().Add(time.Hour).Unix(),
		"hwid": camera.PanelHWID,
		"iat":  time.Now().Unix(),
		"iss":  camera.UUID,
		"uid":  camera.UUID,
	})

	block, _ := pem.Decode(j.Secret)
	if block == nil || block.Type != "RSA PRIVATE KEY" {
		return "", fmt.Errorf("invalid PEM block")
	}
	secret, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return "", err
	}

	result, err := token.SignedString(secret)
	fmt.Printf("HERE: %q\n", result)
	return result, err
}
