package funds

import (
	"gopkg.in/headzoo/surf.v1"
	"github.com/headzoo/surf/browser"
	"regexp"
	"net/http"
	"encoding/xml"
	"ciscowx/ciscoxml"
)

func authenticate(ecUsername string, ecPassword string) (s *browser.Browser) {
	s = surf.NewBrowser()
	s.Open("https://www.easycall.pl/logowanie.html")
	form, _ := s.Form("form[action='logowanie.html']")
	form.Input("log", ecUsername)
	form.Input("pass", ecPassword)
	form.Submit()
	return s
}

func getFunds(s *browser.Browser) string {
	s.Open("https://www.easycall.pl/moje_konto_podsumowanie.html")
	selection := s.Find(".clb > li:nth-child(3) > p:nth-child(2) > b:nth-child(1)").First()
	re := regexp.MustCompile("\\d+\\.\\d{2}")
	return re.FindString(selection.Text())
}

func MakeFundsHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s := authenticate(
			r.Context().Value("EC_USERNAME").(string),
			r.Context().Value("EC_PASSWORD").(string),
		)
		x := ciscoxml.CiscoIPPhoneText{
			Title: "Available funds",
			Text: getFunds(s) + " PLN",
		}

		bytes, _ := xml.Marshal(x)
		w.Write(bytes)
	})
}