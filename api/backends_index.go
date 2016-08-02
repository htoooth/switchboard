package api

import (
	"encoding/json"
	"net/http"
)

type JSONSerializable interface {
	AsJSON() interface{}
}

var BackendsIndex = func(backends JSONSerializable) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		backendsJSON, err := json.Marshal(backends.AsJSON())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_, err = w.Write(backendsJSON)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})
}
