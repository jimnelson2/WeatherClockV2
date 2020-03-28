package transform

import (
	"strings"

	"github.com/shawntoffel/darksky"
	"golang.org/x/text/language"
	"golang.org/x/text/search"
)

const (
	// Tornado alert
	Tornado string = "tornado"

	// Thunderstorm alert
	Thunderstorm string = "thunderstorm"

	// Warning type
	Warning string = "warning"

	// Watch type
	Watch string = "watch"
)

// Active returns the most interesting alert contained in the forecast
func Active(forecast darksky.ForecastResponse) string {
	alert := WarningType(forecast.Alerts)
	switch alert {
	case Tornado:
		return Tornado
	case Thunderstorm:
		return Thunderstorm
	default:
		return ""
	}
}

// WarningType returns the most severe active alert category
func WarningType(alerts []*darksky.Alert) string {
	// TODO need a lot of testing of this...where do I get some data?
	// because the spec doesn't have enough detail
	// also the time and expires fields might be relevant
	var tornado bool = false
	var thunderstorm bool = false
	s := search.New(language.English, search.IgnoreCase)
	for _, a := range alerts {
		// Only interested in warnings
		if strings.EqualFold(a.Severity, Warning) {

			start, _ := s.IndexString(a.Title, Tornado)
			if start >= 0 {
				tornado = true
			}

			start, _ = s.IndexString(a.Title, Thunderstorm)
			if start >= 0 {
				thunderstorm = true
			}
		}
	}
	if tornado {
		return Tornado
	}
	if thunderstorm {
		return Thunderstorm
	}
	return ""

}
