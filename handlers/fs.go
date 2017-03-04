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
	PathVar  string
}

func (f *Fs) Handle(ctx *roboot.Context) {
	if ctx.Req.Method != roboot.MethodGet {
		ctx.Status(http.StatusMethodNotAllowed)
		return
	}

	var path string
	if f.Static {
		path = f.Path
	} else {
		path = filepath.Join(f.Path, ctx.ParamValue(f.PathVar))
	}

	serveFile := f.AllowDir
	if !serveFile {
		stat, err := os.Stat(path)
		if err != nil {
			if os.IsNotExist(err) {
				ctx.Status(http.StatusNotFound)
			} else {
				ctx.Env().Error.Handle(ctx, http.StatusInternalServerError, err)
			}
			return
		}
		serveFile = !stat.IsDir()
	}

	if serveFile {
		http.ServeFile(ctx.Resp, ctx.Req, path)
	} else {
		ctx.Status(http.StatusNotFound)
	}
}
