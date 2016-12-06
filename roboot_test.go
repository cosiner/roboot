package roboot_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cosiner/roboot"
	"github.com/cosiner/roboot/router"
)

func TestRoboot(t *testing.T) {
	s := roboot.NewServer(&roboot.Env{}, router.New())

	r := s.Router("")
	r.Handle("/*path", roboot.HandlerFunc(func(ctx *roboot.Context) {
		ctx.Resp.Write([]byte(ctx.Params.Get("path")))
	}))
	r.Handle("/user/:id/info", roboot.HandlerFunc(func(ctx *roboot.Context) {
		ctx.Resp.Write([]byte(ctx.Params.Get("id")))
	}))

	req, _ := http.NewRequest("GET", "/user/id/info", nil)
	recorder := httptest.NewRecorder()

	s.ServeHTTP(recorder, req)

	if recorder.Body.String() != "id" {
		t.Fatal("process failed")
	}

	req, _ = http.NewRequest("GET", "/user/id/inf", nil)
	recorder = httptest.NewRecorder()

	s.ServeHTTP(recorder, req)
	if recorder.Body.String() != "user/id/inf" {
		t.Fatal("process failed")
	}
}
