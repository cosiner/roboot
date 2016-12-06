package roboot

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRoboot(t *testing.T) {
	s := NewServer(&Env{
		Errorf:              t.Errorf,
		FileUploadMaxMemory: 512 << 10,
	})

	r := s.Router("")
	r.Handle("/*path", HandlerFunc(func(ctx *Context) {
		ctx.Resp.Write([]byte(ctx.Params.Get("path")))
	}))
	r.Handle("/user/:id/info", HandlerFunc(func(ctx *Context) {
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
