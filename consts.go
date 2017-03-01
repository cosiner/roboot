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
	HeaderCorsRequestmethod  = "Access-Control-Request-Method"
	HeaderCorsRequestheaders = "Access-Control-Request-Headers"

	HeaderCorsAlloworigin      = "Access-Control-Allow-Origin"
	HeaderCorsAllowcredentials = "Access-Control-Allow-Credentials"
	HeaderCorsAllowheaders     = "Access-Control-Allow-Headers"
	HeaderCorsAllowmethods     = "Access-Control-Allow-Methods"
	HeaderCorsExposeheaders    = "Access-Control-Expose-Headers"
	HeaderCorsMaxage           = "Access-Control-Max-Age"
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
