package robiptool

import (
  "log"
  "time"

  "github.com/andlabs/ui"
)

func showUI() {
  binding := &Binding{}

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

func newPortCombobox(binding *Binding, ports []string) *ui.Combobox {
  portCombobox := ui.NewCombobox()
  binding.PortCombobox = portCombobox

  for _, port := range ports {
    portCombobox.Append(port)
  }
  portCombobox.SetSelected(0)
  return portCombobox
}

func updatePorts(binding *Binding, box *ui.Box) {
  for {
    time.Sleep(500 * time.Millisecond)
    if ports, err := Ports(); err == nil {
      if len(ports) != binding.CurrentPorts {
        binding.CurrentPorts = len(ports)
        ui.QueueMain(func() {
          box.Delete(1)
          box.Append(newPortCombobox(binding, ports), false)
        })
      }
    }
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
  if ports, err := Ports(); err == nil {
    inputField2Box.Append(newPortCombobox(binding, ports), false)
    binding.CurrentPorts = len(ports)

    go updatePorts(binding, inputField2Box)
  }

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
    if ports, err := Ports(); err == nil {
      go start(robipIdField.Text(), ports[binding.PortCombobox.Selected()], binding)
    }
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

    // if err := WriteDataToPort(file.Name(), port, UpdateProgressFn(binding.ProgressBar)); err != nil {
    if err := WriteByEsptool(file.Name(), port, UpdateProgressFn(binding.ProgressBar)); err != nil {
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
  CurrentPorts int
  PortCombobox *ui.Combobox
  PortIndex int
  RobipID string
  LogMessage string
  IsWriting bool
  Window *ui.Window
  MessageLabel *ui.Label
  ProgressBar *ui.ProgressBar
}

func (binding *Binding) LogMessages() string {
  return binding.LogMessage
}

func (binding *Binding) AddMessage(message string) {
  binding.LogMessage += message
}

func (binding *Binding) OnSelectPort(index int) {
  binding.PortIndex = index
}
