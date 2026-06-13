package constants

// Error messages
const (
	ErrMsgInternalServerError = "An unexpected error has occurred. Please try again later"
	ErrMsgFileSystemIsNil     = "filesystem is nil"
	ErrMsgHtmxNotSupported    = "HTMX requests are not supported on this route"
	ErrMsgUnauthorised        = "unable to authorise user with bearer token"
)

// Error message prefixes
const (
	ErrMsgPrefixMissingTemplateFile = "template file not found: "
	ErrMsgPrefixMissingTemplateData = "template data not found: "
)

// Slog keys
const (
	SlogKeyRequestMethod               = "request_method"
	SlogKeyRequestPath                 = "request_path"
	SlogKeyRequestTime                 = "request_time"
	SlogKeyRequestPanicMsg             = "request_panic_msg"
	SlogKeyRequestPanicStack           = "request_panic_stack"
	SlogKeyRequestDurationMs           = "request_duration_ms"
	SlogKeyRequestErr                  = "request_error"
	SlogKeyResponseWriterErr           = "response_writer_error"
	SlogKeyResponseStatusCode          = "response_status_error"
	SlogKeyNonFatalErrParseWithClaims  = "non_fatal_err_parse_with_claims"
	SlogKeyNonFatalErrClaimsGetSubject = "non_fatal_err_claims_get_subject"
	SlogKeyNonFatalErrRetrieveUserById = "non_fatal_err_retrieve_user_by_id"
)

// Other
const (
	IdentifierJwtCookie   = "hg_login_jwt"
	DefaultAppErrorString = "default app error"
)
