package model

import (
	"strings"
	"time"
)

const (
	ScopeDriveFile = "https://www.googleapis.com/auth/drive.file"
	ScopeDrive     = "https://www.googleapis.com/auth/drive"
)

// OauthTokenはOAuthトークンを表す構造体
type OauthToken struct {
	ExpiresAt time.Time // トークンの有効期限
	Scopes    []string  // トークンに付与されているスコープ
}

// IsExpiredはトークンが期限切れかどうかを判定するメソッド
func (t *OauthToken) IsExpired() bool {
	return time.Now().After(t.ExpiresAt)
}

// CanWriteDriveはトークンがGoogle Driveに書き込み可能かどうかを判定するメソッド
func (t *OauthToken) CanWriteDrive() bool {
	for _, scope := range t.Scopes {
		if scope == ScopeDriveFile || scope == ScopeDrive {
			return true
		}
	}
	return false
}

// parseScopesはスコープ文字列を解析してスライスに変換する関数
func parseScopes(scope string) []string {
	if len(scope) == 0 {
		return []string{}
	}
	// 文字列を空白で分割してスライスに変換(OAuth2の使用で、スコープは空白で区切られる)
	return strings.Fields(scope)
}
