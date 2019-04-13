package display

type Color struct {
	R, G, B uint8
}

func (Color) Black() Color {
	return Color{R: 0, G: 0, B: 0}
}

func (Color) DarkBlue() Color {
	return Color{R: 0, G: 0, B: 255}
}

func (Color) Green() Color {
	return Color{R: 0, G: 255, B: 0}
}

func (Color) LightBlue() Color {
	return Color{R: 0, G: 255, B: 255}
}

func (Color) Orange() Color {
	return Color{R: 255, G: 127, B: 0}
}

func (Color) Pink() Color {
	return Color{R: 255, G: 0, B: 255}
}

func (Color) Purple() Color {
	return Color{R: 127, G: 0, B: 255}
}

func (Color) Red() Color {
	return Color{R: 255, G: 0, B: 0}
}

func (Color) Yellow() Color {
	return Color{R: 255, G: 255, B: 0}
}
