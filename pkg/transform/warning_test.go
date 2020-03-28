package transform

import (
	"testing"

	"github.com/shawntoffel/darksky"
	"github.com/stretchr/testify/assert"
)

// TODO we need test setup/teardown to consolidate the test data
func TestWarningTypeEmpty(t *testing.T) {
	result := WarningType(nil)

	if result != "" {
		assert.Empty(t, result, "result should be empty")
	}
}

func TestWarningTypeTornado(t *testing.T) {
	data := make([]*darksky.Alert, 3)
	data[0] = &darksky.Alert{
		Severity: "watch",
		Title:    "Tornado watch",
	}
	data[1] = &darksky.Alert{
		Severity: "warning",
		Title:    "Tornado warning",
	}
	data[2] = &darksky.Alert{
		Severity: "warning",
		Title:    "Thunderstorm warning",
	}

	result := WarningType(data)

	assert.Equal(t, "tornado", result)
}

func TestWarningTypeThunderstorm(t *testing.T) {
	data := make([]*darksky.Alert, 3)
	data[0] = &darksky.Alert{
		Severity: "watch",
		Title:    "Tornado watch",
	}
	data[1] = &darksky.Alert{
		Severity: "watch",
		Title:    "Thunderstorm watch",
	}
	data[2] = &darksky.Alert{
		Severity: "warning",
		Title:    "Thunderstorm warning",
	}

	result := WarningType(data)

	assert.Equal(t, "thunderstorm", result)
}

func TestActiveTornado(t *testing.T) {
	expected := "tornado"

	fr := darksky.ForecastResponse{}
	fr.Alerts = make([]*darksky.Alert, 3)
	fr.Alerts[0] = &darksky.Alert{
		Severity: "watch",
		Title:    "Tornado watch",
	}
	fr.Alerts[1] = &darksky.Alert{
		Severity: "warning",
		Title:    "Tornado warning",
	}
	fr.Alerts[2] = &darksky.Alert{
		Severity: "warning",
		Title:    "Thunderstorm warning",
	}

	result := Active(fr)

	assert.Equal(t, expected, result)
}

func TestActiveThunderstorm(t *testing.T) {
	expected := "thunderstorm"

	fr := darksky.ForecastResponse{}
	fr.Alerts = make([]*darksky.Alert, 3)
	fr.Alerts[0] = &darksky.Alert{
		Severity: "watch",
		Title:    "Tornado watch",
	}
	fr.Alerts[1] = &darksky.Alert{
		Severity: "watch",
		Title:    "Thunderstorm watch",
	}
	fr.Alerts[2] = &darksky.Alert{
		Severity: "warning",
		Title:    "Thunderstorm warning",
	}

	result := Active(fr)

	assert.Equal(t, expected, result)
}

func TestActiveNothing(t *testing.T) {
	expected := ""

	fr := darksky.ForecastResponse{}
	fr.Alerts = make([]*darksky.Alert, 2)
	fr.Alerts[0] = &darksky.Alert{
		Severity: "watch",
		Title:    "Tornado watch",
	}
	fr.Alerts[1] = &darksky.Alert{
		Severity: "watch",
		Title:    "Thunderstorm watch",
	}

	result := Active(fr)

	assert.Equal(t, expected, result)
}
