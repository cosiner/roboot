package roboot_test

import (
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cosiner/roboot"
	"github.com/cosiner/roboot/codec"
	"github.com/cosiner/roboot/router"
)

type errorHandler struct{}

func (errorHandler) Log(ctx *roboot.Context, errType roboot.ErrType, err error) {
	log.Println("handle failed:", errType.String(), err)
}

func (errorHandler) Handle(ctx *roboot.Context, callerDepth int, status int, err error) {
	ctx.Status(status)
	log.Println("handle failed:", err)
}

func TestRoboot(t *testing.T) {
	s := roboot.NewServer(roboot.Env{Codec: codec.JSON, Error: errorHandler{}}, router.New())

	r := s.Router("")
	r.Handle("/*path", roboot.HandlerFunc(func(ctx *roboot.Context) {
		ctx.Resp.Write([]byte(ctx.ParamValue("path")))
	}))
	r.Handle("/user/:id/info", roboot.HandlerFunc(func(ctx *roboot.Context) {
		ctx.Resp.Write([]byte(ctx.ParamValue("id")))
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
