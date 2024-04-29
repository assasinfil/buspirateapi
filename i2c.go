package buspirateapi

import (
	"errors"
	"time"
)

const (
	V3V = 0x01
	V5V = 0x02
)

const (
	I2C_CS      = 0x01
	I2C_AUX     = 0x02
	I2C_PULLUPS = 0x04
	I2C_POWER   = 0x08
)

const (
	S5KHZ   = 0x00
	S50KHZ  = 0x01
	S100KHZ = 0x02
	S400KHZ = 0x03
)

const (
	COMMAND_EXIT  = 0x00 //Exit to bitbang mode, responds "BBIOx"
	COMMAND_MODE  = 0x01 //Display mode version string, responds "I2Cx"
	COMMAND_START = 0x02 //I2C start bit
	COMMAND_STOP  = 0x03 //I2C stop bit
	COMMAND_READ  = 0x04 //I2C read byte
	COMMAND_ACK   = 0x06 //ACK bit
	COMMAND_NACK  = 0x07 //NACK bit
	COMMAND_SNIFF = 0x0F //Start bus sniffer
	COMMAND_WRITE = 0x10 //Bulk I2C write, send 1-16 bytes (0=1byte!)
	COMMAND_CONF  = 0x40 //Configure peripherals w=power, x=pullups, y=AUX, z=CS
	COMMAND_PULL  = 0x50 //Pull up voltage select (BPV4 only)- x=5v y=3.3v
	COMMAND_SPEED = 0x60 //Set I2C speed, 3=~400kHz, 2=~100kHz, 1=~50kHz, 0=~5kHz
	COMMAND_WTR   = 0x08 //Write then read
)

type I2C struct {
	bp *BusPirate
}

func NewI2C(bp *BusPirate, params []byte) *I2C {
	out := &I2C{bp}
	err := out.SetSpeed(params[0])
	if err != nil {
		return nil
	}
	err = out.ConfigPins(params[1])
	if err != nil {
		return nil
	}
	err = out.SetPullUpVoltage(params[2])
	if err != nil {
		return nil
	}
	return out
}

func (v *I2C) SetSpeed(speed byte) error {
	err := v.bp.SendCommand(COMMAND_SPEED | speed)
	if err != nil {
		return err
	}
	res, err := v.bp.ReadResponse(1)
	if err != nil {
		return err
	}
	if res[0] != 0x01 {
		return errors.New("I2C speed not set")
	}
	return nil
}

func (v *I2C) SetPullUpVoltage(volts byte) error {
	err := v.bp.SendCommand(COMMAND_PULL | volts)
	if err != nil {
		return err
	}
	res, err := v.bp.ReadResponse(1)
	if err != nil {
		return err
	}
	if res[0] != 0x01 {
		return errors.New("pull-up voltage not set")
	}
	return nil

}

func (v *I2C) ConfigPins(pins byte) error {
	err := v.bp.SendCommand(COMMAND_CONF | pins)
	if err != nil {
		return err
	}
	res, err := v.bp.ReadResponse(1)
	if err != nil {
		return err
	}
	if res[0] != 0x01 {
		return errors.New("failed to configure pins")
	}
	return nil
}

func (v *I2C) sendStart() error {
	err := v.bp.SendCommand(COMMAND_START)
	if err != nil {
		return err
	}
	res, err := v.bp.ReadResponse(1)
	if err != nil {
		return err
	}
	if res[0] != 0x01 {
		return errors.New("failed to send start")
	}
	return nil
}

func (v *I2C) sendStop() error {
	err := v.bp.SendCommand(COMMAND_STOP)
	if err != nil {
		return err
	}
	res, err := v.bp.ReadResponse(1)
	if err != nil {
		return err
	}
	if res[0] != 0x01 {
		return errors.New("failed to send stop")
	}
	return nil
}

func (v *I2C) write(data []byte) error {
	if data == nil {
		return errors.New("data is nil")
	}
	if len(data) > 0xF {
		return errors.New("data is too long")
	}
	if err := v.bp.SendCommand(COMMAND_WRITE | byte(len(data)-1)); err != nil {
		return err
	}
	time.Sleep(10 * time.Millisecond)
	if val, err := v.bp.ReadResponse(1); err != nil || val[0] != 0x01 {
		return errors.New("failed to start writing")
	}
	if _, err := v.bp.device.Write(data); err != nil {
		return err
	}
	time.Sleep(10 * time.Millisecond)
	if _, err := v.bp.ReadResponse(len(data)); err != nil {
		return errors.New("failed to write data")

	}
	return nil
}

func (v *I2C) Write(data []byte) error {
	if data == nil {
		return errors.New("data is nil")
	}
	if len(data) > 0xF {
		return errors.New("data is too long")
	}
	if err := v.sendStart(); err != nil {
		return err
	}
	if err := v.write(data); err != nil {
		return err
	}
	if err := v.sendStop(); err != nil {
		return err
	}
	return nil

}

func (v *I2C) readByte() (byte, error) {
	_, err := v.bp.device.Write([]byte{0x04})
	if err != nil {
		return 0, err
	}
	time.Sleep(10 * time.Millisecond)
	resp, err := v.bp.ReadResponse(1)
	return resp[0], err
}

func (v *I2C) Read(address byte, data []byte) (int, error) {
	if err := v.sendStart(); err != nil {
		return 0, err
	}
	if err := v.write([]byte{address}); err != nil {
		return 0, err
	}
	size := len(data)
	n := 0
	for size > 0 {
		b, _ := v.readByte()
		data[n] = b
		if size > 1 {
			_, _ = v.sendAck()
		}
		size--
		n++
	}
	_, _ = v.sendNAck()
	if err := v.sendStop(); err != nil {
		return 0, err
	}
	return n, nil
}

func (v *I2C) sendAck() ([]byte, error) {
	_, err := v.bp.device.Write([]byte{COMMAND_ACK})
	if err != nil {
		return nil, err
	}
	return v.bp.ReadResponse(1)
}

func (v *I2C) sendNAck() ([]byte, error) {
	_, err := v.bp.device.Write([]byte{COMMAND_NACK})
	if err != nil {
		return nil, err
	}
	return v.bp.ReadResponse(1)
}
