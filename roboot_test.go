package roboot

import "testing"

func TestRoboot(t *testing.T) {
	s := NewServer()
	r := s.Router("")
	r.Handle("/*path", HandlerFunc(NotFoundHandler))
	t.Log(r.MatchHandler("/path"))
}
