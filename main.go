package main

import (
	"net/http"
	"encoding/xml"
	"ciscowx/ciscoxml"
	"ciscowx/funds"
	"ciscowx/weather"
)


func rootHandler(w http.ResponseWriter, r *http.Request) {
	x := ciscoxml.CiscoIPPhoneMenu{
		Title: "Exciting services",
		MenuItem: []ciscoxml.MenuItem{
			ciscoxml.MenuItem{
				Name: "Weather",
				URL: "/wx",
			},
			ciscoxml.MenuItem{
				Name: "Available funds",
				URL: "/funds",
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
	http.Handle("/funds", contentTypeXml(funds.MakeFundsHandler()))
	http.Handle("/wx", contentTypeXml(weather.MakeWxRootHandler()))
	http.Handle("/wx/", contentTypeXml(weather.MakeWxHandler()))
	http.Handle("/", contentTypeXml(http.HandlerFunc(rootHandler)))
	http.ListenAndServe("0.0.0.0:8080", nil)
}
