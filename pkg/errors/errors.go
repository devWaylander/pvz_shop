package errors

const (
	// ===================-  COMMON  -===================
	ErrDecodeCtx = "ERR_FAILED_TO_DECODE_CONTEXT_CLAIMS"
	// ===================-  USER  -===================
	ErrUserNotFound = "ERR_USER_NOT_FOUND"
	ErrUserExist    = "ERR_USER_ALREADY_EXIST"
	// ===================-  AUTH  -===================
	ErrEncodeJWT           = "ERR_FAILED_TO_ENCODE_JWT"
	ErrGenUUID             = "ERR_FAILED_TO_GEN_RANDOM_UUID"
	ErrWrongPassword       = "ERR_WRONG_PASSWORD"
	ErrWrongPasswordFormat = "ERR_WRONG_PASSWORD_FORMAT"
	ErrInvalidToken        = "ERR_INVALID_AUTH_TOKEN"
	ErrInvalidClaims       = "ERR_CANNOT_PARSE_CLAIMS"
	ErrUnauthenticated     = "ERR_UNAUTHENTICATED"
)
