package robiptool

import (
  "log"

  "github.com/andlabs/ui"
)

func showUI() {
  binding := &Binding{}
  if ports, err := Ports(); err == nil {
    for _, port := range ports {
      binding.AddPort(port)
    }
  }

  if err := ui.Main(func() {
    window := ui.NewWindow("Robip tool ver 1.2", 200, 40, false)

    binding.Window = window

    window.SetChild(components(binding))
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

func components(binding *Binding) *ui.Box {
  mainBox := ui.NewVerticalBox()
  mainBox.SetPadded(true)

  inputField1Box := ui.NewHorizontalBox()
  inputField2Box := ui.NewHorizontalBox()

  inputField1Box.Append(ui.NewLabel("Robip ID: "), false)

  robipIdField := ui.NewEntry()
  inputField1Box.Append(robipIdField, false)

  inputField2Box.Append(ui.NewLabel("ポート: "), false)
  portCombobox := ui.NewCombobox()
  for _, port := range binding.Ports {
    portCombobox.Append(port)
  }
  portCombobox.SetSelected(0)
  inputField2Box.Append(portCombobox, false)

  button := ui.NewButton("書き込む!")
  inputField1Box.Append(button, false)

  mainBox.Append(inputField1Box, false)
  mainBox.Append(inputField2Box, false)

  mainBox.Append(ui.NewHorizontalSeparator(), false)

  progress := ui.NewProgressBar()
  binding.ProgressBar = progress

  progress.SetValue(0)
  mainBox.Append(progress, true)

  message := ui.NewLabel("")
  binding.MessageLabel = message

  mainBox.Append(message, true)

  button.OnClicked(func(*ui.Button) {
    go start(robipIdField.Text(), binding.PortAt(portCombobox.Selected()), binding)
  })

  return mainBox
}

func doneWriting(binding *Binding, message string, err error) {
  binding.IsWriting = false
  if err != nil {
    log.Println(err)
  }
  go ui.QueueMain(func() {
    binding.ProgressBar.SetValue(0)
    binding.MessageLabel.SetText("")
    ui.MsgBox(binding.Window, "書き込み", message)
  })
}

func updateProgress(binding *Binding, progressVal int, message string) {
  go ui.QueueMain(func() {
    binding.ProgressBar.SetValue(progressVal)
    binding.MessageLabel.SetText(message)
  })
}

func start(id string, port string, binding *Binding) {
  if binding.IsWriting {
    return
  }

  binding.IsWriting = true

  updateProgress(binding, 0, "ファイルの読込中...")

  if file, err := FetchBinary(id); err != nil {
    doneWriting(binding, "Robip IDからファイルを取得できませんでした", err)

  } else {
    updateProgress(binding, 10, "書き込み中...")

    if err := WriteDataToPort(file.Name(), port, UpdateProgressFn(binding.ProgressBar)); err != nil {
      doneWriting(binding, "書き込みに失敗しました", err)

    } else {
      doneWriting(binding, "書き込みに成功しました!", nil)
    }
  }
}

func UpdateProgressFn(progressBar *ui.ProgressBar) UpdateProgress {
  return func (value float32) {
    ui.QueueMain(func() {
      progressBar.SetValue(10 + int(value * 0.9))
    })
  }
}

type Binding struct {
	Ports []string
  PortIndex int
	Counter	int
  RobipID string
  LogMessage string
  IsWriting bool
  Window *ui.Window
  MessageLabel *ui.Label
  ProgressBar *ui.ProgressBar
}

func (binding  *Binding) AddPort(port string) {
	binding.Ports = append(binding.Ports, port)
	binding.Counter++
}

func (binding *Binding) LogMessages() string {
  return binding.LogMessage
}

func (binding *Binding) AddMessage(message string) {
  binding.LogMessage += message
	binding.Counter++
}

func (binding *Binding) PortsLength() int {
	return len(binding.Ports)
}

func (binding *Binding) PortAt(index int) string {
	return binding.Ports[index]
}

func (binding *Binding) OnSelectPort(index int) {
  binding.PortIndex = index
}

func (binding *Binding) OnClicked() {
  port := binding.PortAt(binding.PortIndex)
  log.Printf("Robip ID: %v, Port: %v\n", binding.RobipID, port)
}
