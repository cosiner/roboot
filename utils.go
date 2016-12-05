package roboot

import "net/http"

func NotFoundHandler(req Request, res Response) {
	res.Status(http.StatusNotFound)
}
