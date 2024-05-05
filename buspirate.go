package buspirate

import (
	"errors"
	"fmt"
	"go.bug.st/serial"
	"time"
)

const (
	BP_CS     = 0x01
	BP_MISO   = 0x02
	BP_CLK    = 0x04
	BP_MOSI   = 0x08
	BP_AUX    = 0x16
	BP_PULLUP = 0x32
	BP_POWER  = 0x64
)

const (
	commandReset   = 0x00 //Reset, responds "BBIO1"
	commandSpi     = 0x01 //Enter binary SPI mode, responds "SPI1"
	commandI2C     = 0x02 //Enter binary I2C mode, responds "I2C1"
	commandUart    = 0x03 //Enter binary UART mode, responds "ART1"
	command1WIRE   = 0x04 //Enter binary 1-Wire mode, responds "1W01"
	commandRaw     = 0x05 //Enter binary raw-wire mode, responds "RAW1"
	commandJtag    = 0x06 //Enter OpenOCD JTAG mode
	commandResetbp = 0x0F //Reset Bus Pirate
	commandStest   = 0x10 //Bus Pirate self-test short
	commandLtest   = 0x11 //Bus Pirate self-test long
	commandPwm     = 0x12 //Setup pulse-width modulation (requires 5 byte setup)
	commandCpwm    = 0x13 //Clear/disable PWM
	commandVpm     = 0x14 //Take voltage probe measurement (returns 2 bytes)
	commandCvpm    = 0x15 //Continuous voltage probe measurement
	commandFma     = 0x16 //Frequency measurement on AUX pin
	commandPinsio  = 0x40 //Configure pins as input(1) or output(0): AUX|MOSI|CLK|MISO|CS
	commandPinspw  = 0x80 //Set on (1) or off (0): POWER|PULLUP|AUX|MOSI|CLK|MISO|CS
)

const (
	BpSpiMode     = commandSpi
	BpI2CMode     = commandI2C
	BpUartMode    = commandUart
	BpOneWireMode = command1WIRE
	BpRawMode     = commandRaw
	BpJTAGMode    = commandJtag
)

type Transport interface {
	Write([]byte) error
	Read(byte, []byte) (int, error)
}

type BusPirate struct {
	device    serial.Port
	Transport Transport
}

func NewBusPirate(port serial.Port) *BusPirate {
	return &BusPirate{device: port}
}

func (b *BusPirate) Connect() error {
	err := b.device.SetReadTimeout(500 * time.Millisecond)
	if err := b.device.ResetInputBuffer(); err != nil {
		return err
	}
	if err := b.device.ResetOutputBuffer(); err != nil {
		return err
	}
	if err != nil {
		return err
	}
	fmt.Print("Connecting to Bus Pirate")
	for i := 0; i < 30; i++ {
		fmt.Print(".")
		if err := b.SendCommand(commandReset); err == nil {
			res, err := b.ReadResponse(5)
			if err == nil {
				if string(res) == "BBIO1" {
					fmt.Println("\nConnected")
					return nil
				}
			}
		}
	}

	return nil
}

func (b *BusPirate) SendCommand(command byte) error {
	//log.Println("Sending command")
	_, err := b.device.Write([]byte{command})
	if err != nil {
		return err
	}
	return nil
}

func (b *BusPirate) ReadResponse(size int) ([]byte, error) {
	//log.Println("Reading response")
	buf := make([]byte, size)
	count, err := b.device.Read(buf)
	if err != nil {
		return nil, err
	}
	if count == 0 {
		return nil, errors.New("no data read")
	}
	for count < size {
		n, err := b.device.Read(buf[count:])
		if err != nil {
			return nil, err
		}
		count += n
	}
	return buf, nil
}

func (b *BusPirate) sendMode(mode byte) error {
	if err := b.SendCommand(mode); err == nil {
		res, err := b.ReadResponse(4)
		if err == nil && string(res) == responseForMode(mode) {
			return nil
		}
	}
	return errors.New("failed to send mode")
}

func (b *BusPirate) SwitchMode(mode byte, params []byte) error {
	switch mode {
	case BpI2CMode:
		if err := b.sendMode(commandI2C); err == nil {
			b.Transport = NewI2C(b, params)
		}
	}
	if b.Transport != nil {
		return nil
	} else {
		return errors.New("failed to switch mode")
	}
}

func responseForMode(mode byte) string {
	switch mode {
	case BpSpiMode:
		return "SPI1"
	case BpI2CMode:
		return "I2C1"
	case BpUartMode:
		return "ART1"
	case BpOneWireMode:
		return "1W01"
	case BpRawMode:
		return "RAW1"
	case BpJTAGMode:
		return "JTAG1"
	}
	return "Unknown mode"
}
