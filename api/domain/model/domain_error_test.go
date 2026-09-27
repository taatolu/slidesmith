package model

import (
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

// TestDomainErrorはDomainErrorが意図した通りにErrorをラップして返すか確認する
func TestDomainError(t *testing.T) {
	t.Run("fmt.Errorf で包まれていても errors.As で取り出せる", func(t *testing.T) {
		assert := assert.New(t)
		// DomainErrorに適当な値を設定してdeに格納 -> Errorfでラップ
		de := NewDomainError(ErrCodeUnauthenticated, "期限切れ", nil)
		wrapped := fmt.Errorf("sampleError: %w", de)

		// DomainErrorの構造体をgotに格納
		var got *DomainError
		ok := errors.As(wrapped, &got) //ラップしたErrorをDomainError型に変換できるか確認
		assert.True(ok)
		assert.Equal(ErrCodeUnauthenticated, got.Code)
		assert.Equal("期限切れ", got.Message)
	})
}
