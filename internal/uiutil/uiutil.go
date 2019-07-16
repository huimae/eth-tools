package uiutil

import "github.com/andlabs/ui"

// GetEntry 生成一个输入款
func GetEntry(name string) (et *ui.Entry, box *ui.Box) {
	box = ui.NewHorizontalBox()
	box.SetPadded(true)
	label := ui.NewLabel(name)
	et = ui.NewEntry()
	box.Append(label, false)
	box.Append(et, true)
	return
}
