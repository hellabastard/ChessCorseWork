package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

type CustomButton struct {
	widget.Button
	onRightClick func()
}

func NewCustomButton(label string, onLeftClick, onRightClick func()) *CustomButton {
	b := &CustomButton{
		Button: widget.Button{
			Text:     label,
			OnTapped: onLeftClick,
		},
		onRightClick: onRightClick,
	}
	b.ExtendBaseWidget(b)
	return b
}

func (b *CustomButton) TappedSecondary(*fyne.PointEvent) {
	if b.onRightClick != nil {
		b.onRightClick()
	}
}
