package handlers

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/cosiner/roboot"
)

type Fs struct {
	Static bool
	Path   string

	AllowDir bool
	Pathvar  string
}

func (f *Fs) Handle(ctx *roboot.Context) {
	if ctx.Req.Method != "GET" {
		ctx.Resp.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var path string
	if f.Static {
		path = f.Path
	} else {
		path = filepath.Join(f.Path, ctx.Params.Get(f.Pathvar))
	}

	serveFile := f.AllowDir
	if !serveFile {
		stat, err := os.Stat(path)
		if err != nil {
			if os.IsNotExist(err) {
				ctx.Resp.WriteHeader(http.StatusNotFound)
			} else {
				ctx.Resp.WriteHeader(http.StatusInternalServerError)
				ctx.Env.Errorf("query path stat failed %s: %s", path, err.Error())
			}
			return
		}
		serveFile = !stat.IsDir()
	}

	if serveFile {
		http.ServeFile(ctx.Resp, ctx.Req, path)
	} else {
		ctx.Resp.WriteHeader(http.StatusNotFound)
	}
}
