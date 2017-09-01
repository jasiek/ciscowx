package main

import (
	"github.com/alsm/forecastio"
	"time"
)

type forecastWithTTL struct {
	Forecast *forecastio.Forecast
	ExpiresAt time.Time
}

type Cache struct {
	dict map[complex128]forecastWithTTL
}

type FetchFunction func(lat float64, lng float64) (*forecastio.Forecast)

func NewCache() (*Cache) {
	return &Cache{
		dict: make(map[complex128]forecastWithTTL),
	}
}

func (c *Cache) store(lat float64, lng float64, forecast *forecastio.Forecast) {
	key := complex(lat, lng)
	c.dict[key] = forecastWithTTL{
		Forecast: forecast,
		ExpiresAt: time.Now().Add(8 * time.Hour),
	}
}

func (c *Cache) fetch(lat float64, lng float64,) (*forecastio.Forecast) {
	key := complex(lat, lng)
	if f, ok := c.dict[key]; ok {
		if f.ExpiresAt.Before(time.Now()) {
			delete(c.dict, key)
		} else {
			return f.Forecast
		}
	}
	return nil
}

func (c *Cache) MaybeFetch(lat float64, lng float64, ff FetchFunction) (*forecastio.Forecast) {
	f := c.fetch(lat, lng)
	if f == nil {
		f = ff(lat, lng)
		c.store(lat, lng, f)
	}
	return f
}