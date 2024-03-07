package http

import "net/http"

type FasnowHttp struct {
	client *http.Client
}

func (r *FasnowHttp) Do(request *http.Request) {
	request.Header.Add("User-Agent", "")
}
