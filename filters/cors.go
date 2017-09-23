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
	f.methodsStr = strings.Join(f.methods, ",")

	f.headers = c.Headers
	if len(f.headers) == 0 {
		f.headers = []string{roboot.HeaderOrigin, roboot.HeaderAccept, roboot.HeaderContentType, roboot.HeaderAuthorization}
	}
	for i := range f.headers {
		f.headers[i] = strings.ToLower(f.headers[i]) // chrome browser will use lower header
	}
	f.headersStr = strings.Join(f.headers, ",")

	f.exposeHeadersStr = strings.Join(c.ExposeHeaders, ",")
	if c.AllowCredentials {
		f.allowCredentials = strconv.FormatBool(c.AllowCredentials)
	}

	const defaultPreflightMaxAge = 3600
	if c.PreflightMaxage == 0 {
		c.PreflightMaxage = defaultPreflightMaxAge
	}
	if c.PreflightMaxage > 0 {
		f.preflightMaxage = strconv.Itoa(c.PreflightMaxage)
	}

	return &f
}

func (c *corsFilter) checkOrigin(origin string) bool {
	if origin == "" {
		return true
	}

	var has bool
	for i := 0; i < len(c.origins) && !has; i++ {
		has = c.origins[i] == origin
	}
	return has
}

func (c *corsFilter) setHeaders(ctx *roboot.Context, origin string) {
	headers := ctx.Resp.Header()
	headers.Set(roboot.HeaderCorsAllowOrigin, origin)
	headers.Set(roboot.HeaderCorsAllowMethods, c.methodsStr)
	headers.Set(roboot.HeaderCorsAllowHeaders, c.headersStr)
	if c.allowCredentials != "" {
		headers.Set(roboot.HeaderCorsAllowCredentials, c.allowCredentials)
	}
	if c.exposeHeadersStr != "" {
		headers.Set(roboot.HeaderCorsExposeHeaders, c.exposeHeadersStr)
	}
	if c.preflightMaxage != "" {
		headers.Set(roboot.HeaderCorsMaxAge, c.preflightMaxage)
	}
}

func (c *corsFilter) preflight(ctx *roboot.Context, method, headers, origin string) {
	var allowMethod bool
	for _, m := range c.methods {
		if m == method {
			allowMethod = true
			break
		}
	}
	if !allowMethod {
		ctx.Status(http.StatusForbidden)
		return
	}

	var hdrs []string
	if headers != "" {
		hdrs = strings.Split(headers, ",")
	}
	for _, h := range hdrs {
		h = strings.ToLower(strings.TrimSpace(h))

		var allowHeader bool
		for _, ch := range c.headers {
			if h == ch {
				allowHeader = true
				break
			}
		}
		if !allowHeader {
			ctx.Status(http.StatusForbidden)
			return
		}
	}

	c.setHeaders(ctx, origin)
	ctx.Status(http.StatusOK)
}

func (c *corsFilter) filter(ctx *roboot.Context, chain roboot.Handler, origin string) {
	c.setHeaders(ctx, origin)
	chain.Handle(ctx)
}

func (c *corsFilter) Filter(ctx *roboot.Context, chain roboot.Handler) {
	origin := ctx.Req.Header.Get(roboot.HeaderOrigin)
	if len(c.origins) != 0 && !c.checkOrigin(origin) {
		ctx.Status(http.StatusForbidden)
		return
	}

	reqMethod := ctx.Req.Header.Get(roboot.HeaderCorsRequestMethod)
	reqHeaders := ctx.Req.Header.Get(roboot.HeaderCorsRequestHeaders)
	if ctx.Req.Method == roboot.MethodOptions && (reqMethod != "" || reqHeaders != "") {
		c.preflight(ctx, reqMethod, reqHeaders, origin)
	} else {
		c.filter(ctx, chain, origin)
	}
}
