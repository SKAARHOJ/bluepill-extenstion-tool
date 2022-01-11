package main

import (
	"os"
	"time"

	"bluepill-extenstion-tool/i2c"

	periph "github.com/SKAARHOJ/ibeam-lib-peripherals"
	log "github.com/s00500/env_logger"
)

//pca9500 eeprom i2c adress
var pcaAddrEEPROM uint8 = 0x56

func WriteResourceListToBoard(data periph.EEPROMData) error {
	buf := make([]byte, 7+len(data.Resources)*3)
	// Read metadata
	buf[0] = data.Info.PARAMS
	buf[1] = data.Info.MODEL
	buf[2] = data.Info.DAY
	buf[3] = data.Info.MONTH
	buf[4] = byte(data.Info.YEAR >> 8)
	buf[5] = byte(data.Info.YEAR)
	buf[6] = data.Info.PCB_VER
	// Read resources
	for i, res := range data.Resources {
		off := i*3 + 7 // 3 bytes per entry
		buf[off] = byte(res.Type)
		buf[off+1] = byte(res.IICAddres)
		buf[off+2] = byte(res.BusIndex)
	}
	return pcaWriteBytes(0, buf)
}

var ADLIOIPTX = periph.EEPROMData{
	Info: periph.BoardInfo{
		PARAMS:  14,
		MODEL:   1,
		DAY:     25,
		MONTH:   12,
		YEAR:    2021,
		PCB_VER: 0x01,
	},
	Resources: []periph.Resource{
		{
			Type: periph.ResourceType_Relay,
		},
		{
			Type:     periph.ResourceType_UART_RS485,
			BusIndex: 1,
		},
		{
			Type:     periph.ResourceType_UART_RS422,
			BusIndex: 2,
		},
		{
			Type:     periph.ResourceType_Motor,
			BusIndex: 1,
		},
		{
			Type:     periph.ResourceType_Motor,
			BusIndex: 2,
		},
		{
			Type:     periph.ResourceType_AnalogIn,
			BusIndex: 1,
		},
		{
			Type:     periph.ResourceType_AnalogIn,
			BusIndex: 2,
		},
		{
			Type:     periph.ResourceType_AnalogOut,
			BusIndex: 3,
		},
		{
			Type:     periph.ResourceType_AnalogOut,
			BusIndex: 4,
		},
	},
}

// 250 bytes of eeprom -> 80+ external components possible

func pcaReadBytes(addr byte, bytes int) ([]byte, error) {
	i2cEEPROM, err := i2c.NewI2C(pcaAddrEEPROM, 5)
	if err != nil {
		return nil, log.Wrap(err, "on opening i2c connection")
	}

	buf := make([]byte, bytes)
	log.Debugf("Will try to read to PCA EEPROM at adress : %d , bytes : %X", addr, bytes)

	_, err = i2cEEPROM.WriteBytes([]byte{addr})
	if err != nil {
		return nil, err
	}
	time.Sleep(5 * time.Millisecond)
	_, err = i2cEEPROM.ReadBytes(buf)
	if err != nil {
		return nil, err
	}
	return buf, nil
}

func pcaWriteBytes(addr byte, bytes []byte) error {
	i2cEEPROM, err := i2c.NewI2C(pcaAddrEEPROM, 5)
	if err != nil {
		return log.Wrap(err, "on opening i2c connection")
	}

	log.Debugf("Will try to write to PCA EEPROM at adress : %d , bytes : %X", addr, bytes)
	for i := 0; i < len(bytes); i++ {
		_, err := i2cEEPROM.Write([]byte{byte(addr + byte(i)), bytes[i]})
		if err != nil {
			return err
		}
		time.Sleep(5 * time.Millisecond)
	}
	return nil
}

/*
func ExpansionConnected() bool {
	_, err := i2c.NewI2C(pcaAddrEEPROM, 5)
	if err != nil {
		log.Errorf("Failed to open I2C connection with reason:%s", err)
		return false
	}
	// to do ping some component like pca9500
	return true
}*/

func ReadResourceListFromBoard() (periph.EEPROMData, error) {
	var data periph.EEPROMData
	buf, err := pcaReadBytes(0, 250)
	if err != nil {
		return periph.EEPROMData{}, err
	}

	// Read metadata
	data.Info.PARAMS = buf[0]
	data.Info.MODEL = buf[1]
	data.Info.DAY = buf[2]
	data.Info.MONTH = buf[3]
	data.Info.YEAR = uint16(buf[4])<<8 + uint16(buf[5])
	data.Info.PCB_VER = buf[6]

	// Read resources
	for i := 0; i < 82; i++ {
		off := i*3 + 6 // 3 bytes per entry
		if buf[off] == 0 {
			break
		}
		data.Resources = append(data.Resources, periph.Resource{Type: periph.ResourceType(buf[off]), IICAddres: buf[off+1], BusIndex: buf[off+2]})
	}

	return data, nil
}

func main() {
	log.Info("bluepill-extenstion-tool")
	if len(os.Args) < 2 {
		log.Fatal("Please provide action 'read' or 'write'")
	}

	switch os.Args[1] {
	case "read":
		data, err := ReadResourceListFromBoard()
		log.MustFatal(err)
		log.Info(log.Indent(data))
	case "write":
		// Parse a file!
		err := WriteResourceListToBoard(ADLIOIPTX)
		log.MustFatal(err)
	default:
		log.Error("Invalid command")
	}

}
