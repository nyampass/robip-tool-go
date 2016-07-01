package robiptool

import (
	"flag"
	"fmt"
	"log"

  "github.com/facchinm/go-serial-native"
)

func Run() {
	var isGUI = flag.Bool("gui", false, "GUI mode")
	var port = flag.String("port", "", "Serial port device")
	var _ = flag.Bool("default-port", false, "Use default serial port device")
	var isShowPorts = flag.Bool("ports", false, "Show port devices")
	var binFile = flag.String("file", "", "File path")
	flag.Parse()

	if *isGUI || len(*binFile) == 0 {
    showUI()

	} else if *isShowPorts {
		ports, err := serial.ListPorts()
		if err != nil {
			log.Panic(err)
		}
		log.Printf("Found %d ports:\n", len(ports))
		for _, port := range ports {
			fmt.Println(port.Name())
		}

	} else {
		WriteDataToPort(*binFile, *port, nil)
	}
}
