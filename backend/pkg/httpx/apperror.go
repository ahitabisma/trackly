package httpx

type AppError struct {
	Code   ErrorCode
	Fields map[string]string
	Detail string
}

func (e *AppError) Error() string {
	return string(e.Code)
}
