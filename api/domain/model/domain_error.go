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

// Error は DomainError が error インターフェースを満たすためのメソッドで、Message だけを返す。
// 原因（Err）は内部情報を含みうるので混ぜない。
func (e *DomainError) Error() string {
	return e.Message
}

// Unwrap は原因のエラー（Err）を返す。
// errors.Is / errors.As は Err というフィールド名を知らないので、
// このメソッドを呼んで DomainError の中の原因をたどる。
func (e *DomainError) Unwrap() error {
	return e.Err
}

// NewDomainErrorはDomainErrorの新しいインスタンスを生成するコンストラクタ関数
func NewDomainError(code ErrorCode, msg string, err error) *DomainError {
	return &DomainError{Code: code, Message: msg, Err: err}
}
