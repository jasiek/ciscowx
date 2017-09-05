package main

import (
	"net/http"
	"encoding/xml"
	"ciscowx/ciscoxml"
	"ciscowx/funds"
	"ciscowx/weather"
	"os"
	"context"
	"github.com/abh/geoip"
	"net"
)

const ctxGeoipKey = "geoip"

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

func configurationMiddleware(next http.Handler) http.Handler {
	testing := len(os.Getenv("TESTING")) > 0
	forecastioKey := os.Getenv("FORECASTIO_KEY")
	if len(forecastioKey) == 0 {
		panic("FORECASTIO_KEY not set")
	}
	ecUsername := os.Getenv("EC_USERNAME")
	if len(ecUsername) == 0 {
		panic("EC_USERNAME not set")
	}
	ecPassword := os.Getenv("EC_PASSWORD")
	if len(ecPassword) == 0 {
		panic("EC_PASSWORD is not set")
	}

	return http.HandlerFunc(func (w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), "TESTING", testing)
		ctx = context.WithValue(ctx, "FORECASTIO_KEY", forecastioKey)
		ctx = context.WithValue(ctx, "EC_USERNAME", ecUsername)
		ctx = context.WithValue(ctx, "EC_PASSWORD", ecPassword)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func geoIpMiddleware(next http.Handler) http.Handler {
	gi, _ := geoip.Open("vendor/GeoIPCity.dat")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		testing := r.Context().Value("TESTING").(bool)
		ip, _, _ := net.SplitHostPort(r.RemoteAddr)

		if testing {
			ip = "172.217.20.206"
		}
		record := gi.GetRecord(ip)
		ctx := context.WithValue(r.Context(), ctxGeoipKey, record)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func wrap(next http.Handler) http.Handler {
	return configurationMiddleware(geoIpMiddleware(contentTypeXml(next)))
}

func main() {
	http.Handle("/funds", wrap(funds.MakeFundsHandler()))
	http.Handle("/wx", wrap(weather.MakeWxRootHandler()))
	http.Handle("/wx/", wrap(weather.MakeWxHandler()))
	http.Handle("/", wrap(http.HandlerFunc(rootHandler)))
	http.ListenAndServe("0.0.0.0:8080", nil)
}
