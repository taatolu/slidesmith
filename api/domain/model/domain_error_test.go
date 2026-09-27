package model

import (
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
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
