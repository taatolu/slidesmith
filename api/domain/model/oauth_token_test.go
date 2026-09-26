package model

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

// TestOauthTokenIsExpiredはOauthTokenが期限切れかどうかを判定するメソッドのテスト
func TestOauthTokenIsExpired(t *testing.T) {
	// テーブル駆動テスト
	tests := []struct {
		name     string
		token    *OauthToken
		expected bool
	}{
		{"Expired token", &OauthToken{ExpiresAt: time.Now().Add(-time.Hour)}, true},
		{"Valid token", &OauthToken{ExpiresAt: time.Now().Add(time.Hour)}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := assert.New(t)
			assert.Equal(tt.expected, tt.token.IsExpired())
			//tt.token.IsExpired()のうち、tokenはOauthToken構造体内を指し、IsExpired()はそのメソッドであるため、tokenの有効期限を確認して期限切れかどうかを判定している
		})
	}

}

// TestOauthTokenCanWriteDriveはOauthTokenがGoogleDriveの書き込み権限を持つかを判定するメソッドのテスト
func TestOauthTokenCanWriteDrive(t *testing.T) {
	//table driven test
	tests := []struct {
		name     string
		token    *OauthToken
		expected bool
	}{
		{"Can write drive", &OauthToken{Scopes: []string{"https://www.googleapis.com/auth/drive.file"}}, true},
		{"Can write drive with full access", &OauthToken{Scopes: []string{"https://www.googleapis.com/auth/drive"}}, true},
		{"Can't write drive", &OauthToken{Scopes: []string{"https://www.googleapis.com/auth/drive.readonly"}}, false},
		{"No scopes", &OauthToken{Scopes: []string{}}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := assert.New(t)
			assert.Equal(tt.expected, tt.token.CanWriteDrive())
		})
	}
}

// TestParseScopesはOauthTokenのスコープ解析のテスト
func TestParseScopes(t *testing.T) {
	// table driven test
	tests := []struct {
		name     string
		scope    string
		expected []string
	}{
		{"No scope", "", []string{}},
		{"Single scope", "https://www.googleapis.com/auth/drive.file", []string{"https://www.googleapis.com/auth/drive.file"}},
		{"Multiple scopes", "https://www.googleapis.com/auth/drive.file https://www.googleapis.com/auth/drive", []string{"https://www.googleapis.com/auth/drive.file", "https://www.googleapis.com/auth/drive"}},
		{"double space", "https://www.googleapis.com/auth/drive.file  email", []string{"https://www.googleapis.com/auth/drive.file", "email"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := assert.New(t)
			assert.Equal(tt.expected, parseScopes(tt.scope))
		})
	}
}
