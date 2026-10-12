package model

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestDomainErrorAs は fmt.Errorf で包まれた DomainError を errors.As で取り出せるかを確認
func TestDomainErrorAs(t *testing.T) {
	t.Run("fmt.Errorf で包まれていても errors.As で取り出せる", func(t *testing.T) {
		assert := assert.New(t)
		// DomainErrorに適当な値を設定してdeに格納 -> Errorfでラップ
		de := NewDomainError(ErrCodeUnauthenticated, "期限切れ", nil)
		wrapped := fmt.Errorf("sampleError: %w", de)

		// DomainErrorの構造体をgotに格納
		var got *DomainError
		ok := errors.As(wrapped, &got) // ラップされた*DomainError を探して got に入れる
		assert.True(ok)
		assert.Equal(ErrCodeUnauthenticated, got.Code)
		assert.Equal("期限切れ", got.Message)
	})
}

// TestDomainErrorIs は DomainError が内包するエラーを errors.Is で判定できるかを確認
func TestDomainErrorIs(t *testing.T) {
	t.Run("DomainErrorの中のErrorを errors.Isで取り出せる", func(t *testing.T) {
		assert := assert.New(t)
		// DomainErrorに適当な値を設定してdeに格納
		innerError := errors.New("サンプルエラー")
		de := NewDomainError(ErrCodeUnauthenticated, "期限切れ", innerError)
		// errors.ISでdeからinnerErrorを取り出せるか確認
		ok := errors.Is(de, innerError)
		assert.True(ok)
	})
}
