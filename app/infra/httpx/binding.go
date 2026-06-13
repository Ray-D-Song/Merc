package httpx

import (
	"encoding/json"
	"net/http"

	"github.com/go-playground/form/v4"
)

var formDecoder = form.NewDecoder()

func BindJSON(r *http.Request, v interface{}) error {
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(v)
}

func BindQuery(r *http.Request, v interface{}) error {
	return formDecoder.Decode(v, r.URL.Query())
}
