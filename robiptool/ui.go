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
  portCombobox.Append("更新する..")
  portCombobox.SetSelected(0)
  portCombobox.OnSelected(func(*ui.combobox) {
    if portCombobox.Selected >= len(binding.Ports) {
      updatePorts()
    }
  })
  inputField2Box.Append(portCombobox, false)

  button := ui.NewButton("書き込む!")
  inputField1Box.Append(button, false)

  mainBox.Append(inputField1Box, false)
  mainBox.Append(inputField2Box, false)

  mainBox.Append(ui.NewHorizontalSeparator(), false)

  progress := ui.NewProgressBar()
  progress.Hide()
  progress.SetValue(0)
  mainBox.Append(progress, true)

  message := ui.NewLabel("")
  mainBox.Append(message, true)

  button.OnClicked(func(*ui.Button) {
    start(robipIdField.Text(), binding.PortAt(portCombobox.Selected()), message, progress)
  })

  return mainBox
}

func start(id string, port string, message *ui.Label, progress *ui.ProgressBar) {
  progress.SetValue(0)
  message.SetText("ファイルの読込中...")

  log.Println(id)
  log.Println(port)
  progress.Show()

  if file, err := FetchBinary(id); err != nil {
    log.Println(err)
    message.SetText( "Robip IDからファイルを取得できませんでした")
    progress.Hide()

  } else {
    message.SetText( "書き込み中...")
    progress.SetValue(10)

    if err := WriteDataToPort(file.Name(), port, UpdateProgressFn(progress)); err != nil {
      log.Println(err)
      message.SetText( "書き込みに失敗しました")
      progress.Hide()

    } else {
      message.SetText( "書き込みに成功しました!")
      progress.SetValue(100)
      
    }

  }
}

func UpdateProgressFn(progressBar *ui.ProgressBar) UpdateProgress {
  return func (value int) {
    progressBar.SetValue(value)
  }
}

type Binding struct {
	Ports []string
  PortIndex int
	Counter	int
  RobipID string
  LogMessage string
  Window *ui.Window
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
