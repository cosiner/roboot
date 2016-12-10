package filters

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/cosiner/roboot"
)

type CORS struct {
	Origins          []string
	Methods          []string
	Headers          []string
	ExposeHeaders    []string // these headers can be accessed by javascript
	PreflightMaxage  int      // max efficient seconds of browser preflight
	AllowCredentials bool
}

type corsFilter struct {
	origins []string

	methods          []string
	methodsStr       string
	headers          []string
	headersStr       string
	exposeHeadersStr string

	preflightMaxage  string
	allowCredentials string
}

func (c *CORS) ToFilter() roboot.Filter {
	var f corsFilter
	if l := len(c.Origins); l > 0 && (l != 1 || c.Origins[0] != "*") {
		f.origins = c.Origins
	}

	f.methods = c.Methods
	if len(f.methods) == 0 {
		f.methods = []string{roboot.METHOD_GET, roboot.METHOD_POST, roboot.METHOD_PATCH, roboot.METHOD_PUT, roboot.METHOD_DELETE}
	}
	f.methodsStr = strings.Join(c.Methods, ",")

	f.headers = c.Headers
	if len(f.headers) == 0 {
		f.headers = []string{roboot.HEADER_ORIGIN, roboot.HEADER_ACCEPT, roboot.HEADER_CONTENT_TYPE, roboot.HEADER_AUTHORIZATION}
	}
	f.headersStr = strings.Join(c.Headers, ",")
	for i := range c.Headers {
		c.Headers[i] = strings.ToLower(c.Headers[i]) // chrome browser will use lower header
	}
	f.exposeHeadersStr = strings.Join(c.ExposeHeaders, ",")
	if c.AllowCredentials {
		f.allowCredentials = strconv.FormatBool(c.AllowCredentials)
	}
	if c.PreflightMaxage != 0 {
		f.preflightMaxage = strconv.Itoa(c.PreflightMaxage)
	}

	return &f
}

func (c *corsFilter) allow(origin string) bool {
	var has bool

	for i := 0; i < len(c.origins) && !has; i++ {
		has = c.origins[i] == origin
	}

	return has
}

func (c *corsFilter) preflight(ctx *roboot.Context, method, headers string) {
	origin := "*"
	if len(c.origins) != 0 {
		origin = ctx.Req.Header.Get(roboot.HEADER_ORIGIN)
		if !c.allow(origin) {
			ctx.Resp.WriteHeader(http.StatusOK)
			return
		}
	}

	respHeaders := ctx.Resp.Header()
	respHeaders.Set(roboot.HEADER_CORS_ALLOWORIGIN, origin)
	upperMethod := strings.ToUpper(method)

	for _, m := range c.methods {
		if m == upperMethod {
			respHeaders.Add(roboot.HEADER_CORS_ALLOWMETHODS, method)
			break
		}
	}

	var hdrs []string
	if headers != "" {
		hdrs = strings.Split(headers, ",")
		for i := range hdrs {
			hdrs[i] = strings.TrimSpace(hdrs[i])
		}
	}
	for _, h := range hdrs {
		for _, ch := range c.headers {
			if strings.ToLower(h) == ch { // c.Headers already ToLowered when Init
				respHeaders.Add(roboot.HEADER_CORS_ALLOWHEADERS, ch)
				break
			}
		}
	}

	respHeaders.Set(roboot.HEADER_CORS_ALLOWCREDENTIALS, c.allowCredentials)
	if c.exposeHeadersStr != "" {
		respHeaders.Set(roboot.HEADER_CORS_EXPOSEHEADERS, c.exposeHeadersStr)
	}

	if c.preflightMaxage != "" {
		respHeaders.Set(roboot.HEADER_CORS_MAXAGE, c.preflightMaxage)
	}

	ctx.Resp.WriteHeader(http.StatusOK)
}

func (c *corsFilter) filter(ctx *roboot.Context, chain roboot.HandlerFunc) {
	headers := ctx.Resp.Header()
	origin := "*"
	if len(c.origins) != 0 {
		origin = ctx.Req.Header.Get(roboot.HEADER_ORIGIN)
		if !c.allow(origin) {
			ctx.Resp.WriteHeader(http.StatusForbidden)
			return
		}
	}
	headers.Set(roboot.HEADER_CORS_ALLOWORIGIN, origin)

	headers.Set(roboot.HEADER_CORS_ALLOWMETHODS, c.methodsStr)
	headers.Set(roboot.HEADER_CORS_ALLOWHEADERS, c.headersStr)

	headers.Set(roboot.HEADER_CORS_ALLOWCREDENTIALS, c.allowCredentials)
	if c.exposeHeadersStr != "" {
		headers.Set(roboot.HEADER_CORS_EXPOSEHEADERS, c.exposeHeadersStr)
	}
	if c.preflightMaxage != "" {
		headers.Set(roboot.HEADER_CORS_MAXAGE, c.preflightMaxage)
	}

	chain(ctx)
}

func (c *corsFilter) Filter(ctx *roboot.Context, chain roboot.HandlerFunc) {
	reqMethod := ctx.Req.Header.Get(roboot.HEADER_CORS_REQUESTMETHOD)
	reqHeaders := ctx.Req.Header.Get(roboot.HEADER_CORS_REQUESTHEADERS)

	if ctx.Req.Method == roboot.METHOD_OPTIONS && (reqMethod != "" || reqHeaders != "") {
		c.preflight(ctx, reqMethod, reqHeaders)
	} else {
		c.filter(ctx, chain)
	}
}
