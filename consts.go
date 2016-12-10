package roboot

const (
	HEADER_ACCEPT           = "Accept"
	HEADER_ACCEPT_ENCODING  = "Accept-Encoding"
	HEADER_CONTENT_ENCODING = "Content-Encoding"
	HEADER_CONTENT_TYPE     = "Content-Type"
	HEADER_CONTENT_LENGTH   = "Content-Length"
	HEADER_USER_AGENT       = "User-Agent"

	HEADER_AUTHORIZATION = "Authorization"

	CONTENT_ENCODING_GZIP    = "gzip"
	CONTENT_ENCODING_DEFLATE = "deflate"

	HEADER_ORIGIN = "Origin"

	// CORS
	HEADER_CORS_REQUESTMETHOD  = "Access-Control-Request-Method"
	HEADER_CORS_REQUESTHEADERS = "Access-Control-Request-Headers"

	HEADER_CORS_ALLOWORIGIN      = "Access-Control-Allow-Origin"
	HEADER_CORS_ALLOWCREDENTIALS = "Access-Control-Allow-Credentials"
	HEADER_CORS_ALLOWHEADERS     = "Access-Control-Allow-Headers"
	HEADER_CORS_ALLOWMETHODS     = "Access-Control-Allow-Methods"
	HEADER_CORS_EXPOSEHEADERS    = "Access-Control-Expose-Headers"
	HEADER_CORS_MAXAGE           = "Access-Control-Max-Age"
)

const (
	METHOD_GET     = "GET"
	METHOD_POST    = "POST"
	METHOD_DELETE  = "DELETE"
	METHOD_PUT     = "PUT"
	METHOD_PATCH   = "PATCH"
	METHOD_HEAD    = "HEAD"
	METHOD_OPTIONS = "OPTIONS"
	METHOD_CONNECT = "CONNECT"
	METHOD_TRACE   = "TRACE"
)
