package buspirate

import (
	"fmt"
	"go.bug.st/serial"
	"log"
	"testing"
)

func TestI2C_Read(t *testing.T) {
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
	bp := NewBusPirate(s)
	err = bp.Connect()
	if err != nil {
		log.Fatalln(err)
		return
	}
	err = bp.SwitchMode(BpI2CMode, []byte{BpS400KHZ, BpI2CPower | BpI2CPullUps, BpV3V})
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
	if data[0] != 0x9 {
		t.Fatal("Expected 0x9, got", data[0])
	}
	err = bp.Transport.Write([]byte{0x20, 0x00, 0x17})
	if err != nil {
		log.Fatalln(err)
		return
	}
	_, err = bp.Transport.Read(0x21, data)
	if data[0] != 0x89 {
		t.Fatal("Expected 0x89, got", data[0])
	}
}
