package main

import (
	"buspirateapi/buspirate"
	"fmt"
	"go.bug.st/serial"
	"log"
)

func main() {
	ports, err := serial.GetPortsList()
	if err != nil {
		log.Fatal(err)
	}
	if len(ports) == 0 {
		log.Fatal("No serial ports found!")
	}
	for _, port := range ports {
		fmt.Printf("Found port: %v\n", port)
	}
	s, err := serial.Open(ports[1], &serial.Mode{BaudRate: 115200})
	if err != nil {
		log.Fatal(err)
	}
	bp := buspirate.NewBusPirate(s)
	err = bp.Connect()
	if err != nil {
		log.Fatalln(err)
		return
	}
	err = bp.SwitchMode(buspirate.I2C_MODE, []byte{buspirate.S400KHZ, buspirate.I2C_POWER | buspirate.I2C_PULLUPS, buspirate.V3V})
	if err != nil {
		log.Fatalln(err)
		return
	}
	err = bp.Transport.Write([]byte{0x20, 0x00, 0x16})
	if err != nil {
		log.Fatalln(err)
		return
	}
	data := make([]byte, 1)
	_, err = bp.Transport.Read(0x21, data)
	log.Println(data)

	err = bp.Transport.Write([]byte{0x20, 0x00, 0x17})
	if err != nil {
		log.Fatalln(err)
		return
	}
	_, err = bp.Transport.Read(0x21, data)
	log.Println(data)
}
