package robiptool

import (
  "os"
  "fmt"
  "gopkg.in/qml.v1"
)

func showUI() {
  if err := qml.Run(run); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
  engine := qml.NewEngine()

  binding := &Binding{}
	engine.Context().SetVar("binding", binding)

  if ports, err := Ports(); err == nil {
    for _, port := range ports {
      binding.AddPort(port)
    }
  }

  if controls, err := engine.LoadFile("robiptool/ui.qml"); err != nil {
    return err
  } else {
    window := controls.CreateWindow(nil)
    window.Show()

    window.Wait()
  }
  return nil
}

type Binding struct {
	Ports []string
	Counter	int
}

func (binding  *Binding) AddPort(port string) {
	binding.Ports = append(binding.Ports, port)
	binding.Counter++
	qml.Changed(binding, &binding.Counter)
}

func (binding *Binding) PortsLength() int {
	return len(binding.Ports)
}

func (binding *Binding) PortAt(index int) string {
	return binding.Ports[index]
}

func (binding *Binding) OnSelectPort(index int) {
	// if index != -1 {
	// 	p := binding.Ports[index]
	// 	// log.Println(p)
	// 	fmt.Println(p)
	// }
}

func (binding *Binding) OnClicked() {
	fmt.Println("clicked")
}
