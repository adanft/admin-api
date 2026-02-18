package logger

const (
	MsgInvalidConfiguration     = "invalid configuration"
	MsgDatabaseInitFailed       = "database initialization failed"
	MsgDatabaseCloseFailed      = "database close failed"
	MsgServerStarted            = "server started"
	MsgServerFailed             = "server failed"
	MsgHTTPRequest              = "http_request"
	MsgPanicRecovered           = "panic_recovered"
	MsgResponseMarshalError     = "response marshal error"
	MsgResponseWriteError       = "response write error"
	MsgResponseFallbackWriteErr = "response fallback write error"
	MsgUserRequestFailed        = "user_request_failed"
	MsgAuthRequestFailed        = "auth_request_failed"
)
