package color

// WCColor represented by RGB
type WCColor struct {
	R, G, B uint8
}

// Display colors
var (
	Black     = WCColor{R: 0, G: 0, B: 0}
	DarkBlue  = WCColor{R: 0, G: 0, B: 255}
	Green     = WCColor{R: 0, G: 255, B: 0}
	LightBlue = WCColor{R: 0, G: 255, B: 255}
	Orange    = WCColor{R: 255, G: 127, B: 0}
	Pink      = WCColor{R: 255, G: 0, B: 255}
	Purple    = WCColor{R: 127, G: 0, B: 255}
	Red       = WCColor{R: 255, G: 0, B: 0}
	Yellow    = WCColor{R: 255, G: 255, B: 0}
)
