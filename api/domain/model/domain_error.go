package model

// ErrorCodeは独自のエラーコードとして設定
type ErrorCode string

const (
	ErrCodeUnauthenticated ErrorCode = "UNAUTHENTICATED"
)

// DomainErrorはドメイン層で発生するエラーを表す構造体
type DomainError struct {
	Code    ErrorCode
	Message string
	Err     error
}

// DomainErrorがerrorインターフェースを実装するようにError()メソッドを定義
func (e *DomainError) Error() string {
	return e.Message
}

// NewDomainErrorはDomainErrorの新しいインスタンスを生成するコンストラクタ関数
func NewDomainError(code ErrorCode, msg string, err error) *DomainError {
	return &DomainError{Code: code, Message: msg, Err: err}
}
