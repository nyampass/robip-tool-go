package robiptool

import (
	"flag"
	"fmt"
	"log"

	"github.com/mikepb/go-serial"
)

func Run() {
	var isGUI = flag.Bool("gui", false, "GUI mode")
	var port = flag.String("port", "", "Serial port device")
	var _ = flag.Bool("default-port", false, "Use default serial port device")
	var isShowPorts = flag.Bool("ports", false, "Show port devices")
	flag.Parse()

	if *isGUI {
		fmt.Println("gui")
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
		writeData(*port)
	}
}
