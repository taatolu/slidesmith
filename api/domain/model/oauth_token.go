package model

import "time"

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
