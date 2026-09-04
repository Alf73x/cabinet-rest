package config

import (
	"strings"
	"testing"
	"time"
)

func TestValidateJWT(t *testing.T) {
	validSecret := strings.Repeat("a", minJWTSecretBytes)

	tests := []struct {
		name    string
		jwt     JWT
		wantErr bool
	}{
		{name: "valid", jwt: JWT{Secret: validSecret, TTL: time.Hour}},
		{name: "empty secret", jwt: JWT{TTL: time.Hour}, wantErr: true},
		{name: "default secret", jwt: JWT{Secret: "change-me", TTL: time.Hour}, wantErr: true},
		{name: "short secret", jwt: JWT{Secret: "too-short", TTL: time.Hour}, wantErr: true},
		{name: "surrounding whitespace", jwt: JWT{Secret: " " + validSecret, TTL: time.Hour}, wantErr: true},
		{name: "zero TTL", jwt: JWT{Secret: validSecret}, wantErr: true},
		{name: "negative TTL", jwt: JWT{Secret: validSecret, TTL: -time.Hour}, wantErr: true},
		{name: "TTL too long", jwt: JWT{Secret: validSecret, TTL: maxJWTTTL + time.Second}, wantErr: true},
		{name: "maximum TTL", jwt: JWT{Secret: validSecret, TTL: maxJWTTTL}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validate(&Config{JWT: tt.jwt})
			if (err != nil) != tt.wantErr {
				t.Fatalf("validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
