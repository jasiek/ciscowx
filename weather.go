package main

import (
	"net/http"
	"encoding/xml"
	"github.com/abh/geoip"
	"net"
	"context"
	"github.com/alsm/forecastio"
	"os"
	"time"
	"regexp"
	"fmt"
)

const ctxGeoipKey = "geoip"



var cache = NewCache()

func geoIpMiddleware(next http.Handler) http.Handler {
	gi, _ := geoip.Open("vendor/GeoIPCity.dat")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip, _, _ := net.SplitHostPort(r.RemoteAddr)

		if len(os.Getenv("TESTING")) > 0 {
			ip = "172.217.20.206"
		}
		record := gi.GetRecord(ip)
		ctx := context.WithValue(r.Context(), ctxGeoipKey, record)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func dateEqual(t1 time.Time, t2 time.Time) bool {
	return t1.Truncate(24 * time.Hour).Equal(t2.Truncate(24 * time.Hour))
}

func forecastHandler(latitude float32, longitude float32, date string) http.Handler {
	connection := forecastio.NewConnection(os.Getenv("FORECASTIO_KEY"))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var ff = func(lat float64, lng float64) (*forecastio.Forecast){
			forecast, _ := connection.Forecast(lat, lng, []string{}, false)
			return forecast
		}

		forecast := cache.MaybeFetch(float64(latitude), float64(longitude), ff)

		text := "<empty>"
		dateTime, _ := time.Parse("2006-01-02", date)
		for _, day := range forecast.Daily.Data {
			thisDay := time.Unix(day.TimeUnix, 0)
			if dateEqual(dateTime, thisDay) {
				text = day.Summary
				text += fmt.Sprintf("%d%% chance of precipitation.", int(day.PrecipitationProbability))
				text += fmt.Sprintf("Temperature: %d-%d °C", int(day.TemperatureMin), int(day.TemperatureMax))
				text += fmt.Sprintf("Humidity: %d%%", int(day.Humidity))
				text += fmt.Sprintf("Pressure: %d hPa", int(day.Pressure))
			}
		}

		x := CiscoIPPhoneText{
			Title: "WX for " + date,
			Text: text,
		}

		bytes, _ := xml.Marshal(x)
		w.Write(bytes)
	})
}

func wxHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	record := ctx.Value(ctxGeoipKey).(*geoip.GeoIPRecord)

	re := regexp.MustCompile("\\d{4}-\\d{2}-\\d{2}")
	date := re.FindString(r.RequestURI)
	forecastHandler(record.Latitude, record.Longitude, date).ServeHTTP(w, r)
}

func wxRootHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	record := ctx.Value(ctxGeoipKey).(*geoip.GeoIPRecord)
	city := record.City
	country := record.CountryName

	x := CiscoIPPhoneMenu{
		Title: "WX for " + city + ", " + country,
		MenuItem: []MenuItem{
			MenuItem{
				Name: city + " today",
				URL: "/wx/" + time.Now().Format("2006-01-02"),
			},
		},
	}

	today := time.Now()

	for i := 1; i < 7; i++ {
		date := today.AddDate(0, 0, i)
		formattedDate := date.Format("2006-01-02")
		item := MenuItem{
			Name: city + " " + formattedDate,
			URL: "/wx/" + formattedDate,
		}
		x.MenuItem = append(x.MenuItem, item)
	}

	bytes, _ := xml.Marshal(x)
	w.Write(bytes)
}

func MakeWxHandler() http.Handler {
	return geoIpMiddleware(http.HandlerFunc(wxHandler))
}

func MakeWxRootHandler() http.Handler {
	return geoIpMiddleware(http.HandlerFunc(wxRootHandler))
}