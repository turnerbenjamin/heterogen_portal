package templates

import "time"

var GMT = mustLoadLocation("Europe/London")

func mustLoadLocation(name string) *time.Location {
	location, err := time.LoadLocation(name)
	if err != nil {
		panic("load timezone " + name + ": " + err.Error())
	}
	return location
}
