package renderer

import (
	"errors"
	"html/template"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/cosiner/roboot"
)

type htmlTemplate struct {
	template *template.Template
}

type HTML struct {
	Pathes        map[string]string // <path, trimPrefix>
	AllowSuffixes []string
	Funcs         template.FuncMap
	Delims        []string
}

func (h HTML) ToRenderer() (roboot.Renderer, error) {
	root := template.New("")
	if len(h.Funcs) > 0 {
		root.Funcs(h.Funcs)
	}
	if l := len(h.Delims); l > 0 {
		if l != 2 {
			return nil, errors.New("illegal template delims")
		}
		root.Delims(h.Delims[0], h.Delims[1])
	}
	if len(h.Pathes) == 0 {
		return nil, errors.New("empty template pathes")
	}
	t := &htmlTemplate{
		template: root,
	}
	for path, prefix := range h.Pathes {
		err := t.loadPath(prefix, path, h.AllowSuffixes...)
		if err != nil {
			return nil, err
		}
	}
	return t, nil
}

func (h *htmlTemplate) isSuffixAllowed(suffix string, suffixes []string) bool {
	if len(suffixes) == 0 {
		return true
	}
	for _, s := range suffixes {
		if s == suffix {
			return true
		}
	}
	return false
}

func (h *htmlTemplate) loadPath(trimPrefix, path string, suffixes ...string) error {
	if trimPrefix != "" {
		var err error
		path, err = filepath.Abs(path)
		if err != nil {
			return err
		}
		trimPrefix, err = filepath.Abs(trimPrefix)
		if err != nil {
			return err
		}
	}

	return filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || !h.isSuffixAllowed(strings.TrimPrefix(".", filepath.Ext(path)), suffixes) {
			return nil
		}

		b, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}

		tmpl := h.template.New(strings.TrimPrefix(path, trimPrefix))
		_, err = tmpl.Parse(string(b))
		return err
	})
}

func (h *htmlTemplate) Render(w io.Writer, name string, v interface{}) error {
	return h.template.ExecuteTemplate(w, name, v)
}
