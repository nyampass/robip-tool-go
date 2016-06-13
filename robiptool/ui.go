package robiptool

import (
    "github.com/andlabs/ui"
)

func showUI() {
  if err := ui.Main(func() {
    window := ui.NewWindow("Robip tool", 400, 200, false)
    window.SetChild(components())
    window.OnClosing(func(*ui.Window) bool {
      ui.Quit()
      return true
    })
    window.SetMargined(true)
    window.Show()

  }); err != nil {
    panic(err)
  }
}

func components() *ui.Box {
  mainBox := ui.NewHorizontalBox()
  mainBox.SetPadded(true)

  fieldBox := ui.NewVerticalBox()
  fieldBox.Append(ui.NewLabel("Robip ID: "), false)
  fieldBox.Append(ui.NewLabel("ポート: "), false)
  mainBox.Append(fieldBox, false)

  inputBox := ui.NewVerticalBox()
  name := ui.NewEntry()
  inputBox.Append(name, false)

  c := ui.NewCombobox()
  c.Append("Hoge")
  inputBox.Append(c, false)

  mainBox.Append(inputBox, false)

  buttonBox := ui.NewVerticalBox()
  button := ui.NewButton("書き込む!")
  buttonBox.Append(button, false)
  refreshButton := ui.NewButton("更新")
  buttonBox.Append(refreshButton, false)

  mainBox.Append(buttonBox, false)

  progress := ui.NewProgressBar()
  progress.SetValue(33)
  mainBox.Append(progress, true)

  return mainBox
}
