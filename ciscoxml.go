package main

type MenuItem struct {
	Name string
	URL string
}

type CiscoIPPhoneMenu struct {
	Title string
	MenuItem []MenuItem
}

type CiscoIPPhoneText struct {
	Title string
	Text string
}