package constants

// Error messages
const (
	ErrMsgInternalServerError = "An unexpected error has occurred. Please try again later"
	ErrMsgFileSystemIsNil     = "filesystem is nil"
	ErrMsgHtmxNotSupported    = "HTMX requests are not supported on this route"
	ErrMsgUnauthorised        = "unable to authorise user with bearer token"
	ErrMissingOIDCCodeParam   = "missing code param in sign-in redirect request"
	ErrMissingOIDCStateParam  = "missing state param in sign-in redirect request"
)

// Error message prefixes
const (
	ErrMsgPrefixMissingTemplateFile = "template file not found: "
	ErrMsgPrefixMissingTemplateData = "template data not found: "
)

// Slog keys
const (
	SlogKeyRequestMethod          = "request_method"
	SlogKeyRequestPath            = "request_path"
	SlogKeyRequestTime            = "request_time"
	SlogKeyRequestPanicMsg        = "request_panic_msg"
	SlogKeyRequestPanicStack      = "request_panic_stack"
	SlogKeyRequestDurationMs      = "request_duration_ms"
	SlogKeyRequestErr             = "request_error"
	SlogKeyResponseWriterErr      = "response_writer_error"
	SlogKeyResponseStatusCode     = "response_status"
	SlogKeyNonFatalErrParseAppJWT = "non_fatal_err_parse_app_jwt"
)

// HX Headers
const (
	HxRequestHeaderRequest   = "HX-Request"
	HxResponseHeaderRedirect = "HX-Redirect"
)

// Other
const (
	IdentifierJwtCookie       = "hg_login_jwt"
	IdentifierOidcStateCookie = "hg_oidc_state"
	EmptyAppErrorString       = "default app error"
)
