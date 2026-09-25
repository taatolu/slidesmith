package model

import "time"

// OauthTokenはOAuthトークンを表す構造体
type OauthToken struct{
	ExpiresAt time.Time // トークンの有効期限
}

// IsExpiredはトークンが期限切れかどうかを判定するメソッド
func (t *OauthToken) IsExpired() bool{
	return time.Now().After(t.ExpiresAt)
}