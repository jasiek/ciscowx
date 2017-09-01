package main

import (
	"net/http"
	"encoding/xml"
)


func rootHandler(w http.ResponseWriter, r *http.Request) {
	x := CiscoIPPhoneMenu{
		Title: "Exciting services",
		MenuItem: []MenuItem{
			MenuItem{
				Name: "Weather",
				URL: "/wx",
			},
		},
	}
	bytes, _ := xml.Marshal(x)
	w.Write(bytes)
}

func contentTypeXml(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/xml")
		next.ServeHTTP(w, r)
	})
}

func main() {
	http.Handle("/wx", contentTypeXml(MakeWxRootHandler()))
	http.Handle("/wx/", contentTypeXml(MakeWxHandler()))
	http.Handle("/", contentTypeXml(http.HandlerFunc(rootHandler)))
	http.ListenAndServe("0.0.0.0:8080", nil)
}
