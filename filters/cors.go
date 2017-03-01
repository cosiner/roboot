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
		f.methods = []string{roboot.MethodGet, roboot.MethodPost, roboot.MethodPatch, roboot.MethodPut, roboot.MethodDelete}
	}
	f.methodsStr = strings.Join(c.Methods, ",")

	f.headers = c.Headers
	if len(f.headers) == 0 {
		f.headers = []string{roboot.HeaderOrigin, roboot.HeaderAccept, roboot.HeaderContentType, roboot.HeaderAuthorization}
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
		origin = ctx.Req.Header.Get(roboot.HeaderOrigin)
		if !c.allow(origin) {
			ctx.Resp.WriteHeader(http.StatusOK)
			return
		}
	}

	respHeaders := ctx.Resp.Header()
	respHeaders.Set(roboot.HeaderCorsAlloworigin, origin)
	upperMethod := strings.ToUpper(method)

	for _, m := range c.methods {
		if m == upperMethod {
			respHeaders.Add(roboot.HeaderCorsAllowmethods, method)
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
				respHeaders.Add(roboot.HeaderCorsAllowheaders, ch)
				break
			}
		}
	}

	respHeaders.Set(roboot.HeaderCorsAllowcredentials, c.allowCredentials)
	if c.exposeHeadersStr != "" {
		respHeaders.Set(roboot.HeaderCorsExposeheaders, c.exposeHeadersStr)
	}

	if c.preflightMaxage != "" {
		respHeaders.Set(roboot.HeaderCorsMaxage, c.preflightMaxage)
	}

	ctx.Resp.WriteHeader(http.StatusOK)
}

func (c *corsFilter) filter(ctx *roboot.Context, chain roboot.HandlerFunc) {
	headers := ctx.Resp.Header()
	origin := "*"
	if len(c.origins) != 0 {
		origin = ctx.Req.Header.Get(roboot.HeaderOrigin)
		if !c.allow(origin) {
			ctx.Resp.WriteHeader(http.StatusForbidden)
			return
		}
	}
	headers.Set(roboot.HeaderCorsAlloworigin, origin)

	headers.Set(roboot.HeaderCorsAllowmethods, c.methodsStr)
	headers.Set(roboot.HeaderCorsAllowheaders, c.headersStr)

	headers.Set(roboot.HeaderCorsAllowcredentials, c.allowCredentials)
	if c.exposeHeadersStr != "" {
		headers.Set(roboot.HeaderCorsExposeheaders, c.exposeHeadersStr)
	}
	if c.preflightMaxage != "" {
		headers.Set(roboot.HeaderCorsMaxage, c.preflightMaxage)
	}

	chain(ctx)
}

func (c *corsFilter) Filter(ctx *roboot.Context, chain roboot.HandlerFunc) {
	reqMethod := ctx.Req.Header.Get(roboot.HeaderCorsRequestmethod)
	reqHeaders := ctx.Req.Header.Get(roboot.HeaderCorsRequestheaders)

	if ctx.Req.Method == roboot.MethodOptions && (reqMethod != "" || reqHeaders != "") {
		c.preflight(ctx, reqMethod, reqHeaders)
	} else {
		c.filter(ctx, chain)
	}
}
