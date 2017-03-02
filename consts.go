package roboot

const (
	HeaderAccept          = "Accept"
	HeaderAcceptEncoding  = "Accept-Encoding"
	HeaderContentEncoding = "Content-Encoding"
	HeaderContentType     = "Content-Type"
	HeaderContentLength   = "Content-Length"
	HeaderUserAgent       = "User-Agent"

	HeaderAuthorization = "Authorization"

	ContentEncodingGzip    = "gzip"
	ContentEncodingDeflate = "deflate"

	HeaderOrigin = "Origin"

	// CORS
	HeaderCorsRequestMethod  = "Access-Control-Request-Method"
	HeaderCorsRequestHeaders = "Access-Control-Request-Headers"

	HeaderCorsAllowOrigin = "Access-Control-Allow-Origin"
	HeaderCorsAllowCredentials = "Access-Control-Allow-Credentials"
	HeaderCorsAllowHeaders     = "Access-Control-Allow-Headers"
	HeaderCorsAllowMethods     = "Access-Control-Allow-Methods"
	HeaderCorsExposeHeaders    = "Access-Control-Expose-Headers"
	HeaderCorsMaxAge           = "Access-Control-Max-Age"
)

const (
	MethodGet     = "GET"
	MethodPost    = "POST"
	MethodDelete  = "DELETE"
	MethodPut     = "PUT"
	MethodPatch   = "PATCH"
	MethodHead    = "HEAD"
	MethodOptions = "OPTIONS"
	MethodConnect = "CONNECT"
	MethodTrace   = "TRACE"
)
