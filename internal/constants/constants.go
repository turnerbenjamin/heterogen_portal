package constants

// Error messages
const (
	ErrMsgInternalServerError                   = "an unexpected error has occurred. Please try again later"
	ErrMsgFileSystemIsNil                       = "filesystem is nil"
	ErrMsgUnauthorised                          = "unable to authorise user with bearer token"
	ErrMissingOIDCCodeParam                     = "missing code param in sign-in redirect request"
	ErrMissingOIDCStateParam                    = "missing state param in sign-in redirect request"
	ErrMissingOIDCStateCookie                   = "missing oidc state cookie in sign-in redirect request"
	ErrMsgHtmxNotSupported                      = "htmx requests are not supported on this route"
	ErrMsgFailedToGenerateRandomString          = "failed to generate random string"
	ErrMsgFailedToGenerateOidcStateValue        = "failed to generate oidc state value"
	ErrMsgFailedToGenerateOidcNonceValue        = "failed to generate oidc nonce value"
	ErrMsgFailedToGenerateOidcCodeVerifierValue = "failed to generate oidc code verifier value"
	ErrMsgInvalidOidcState                      = "oidc state string is invalid"
	ErrMsgOidcStateValueMismatch                = "oidc state value returned from auth provider does not match persisted oidc state"
	ErrMsgOidcNonceValueMismatch                = "oidc nonce value returned from auth provider does not match persisted oidc state"
	ErrMsgUnableToAccessIdToken                 = "unable to access id token following code exchange"
	ErrMsgParseAppJwtErrUnableToReadIdToken     = "unable to read id token from app jwt claims"
	ErrMsgParseAppJwtErrUnableToReadUserId      = "unable to read user id from app jwt claims"
	ErrMsgParseAppJwtErrUnableToReadUserOid     = "unable to read user oid from app jwt claims"
	ErrMsgParseAppJwtErrUnableToReadUserEmail   = "unable to read user email from app jwt claims"
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
	EmptyAppErrorString       = "empty"
)
