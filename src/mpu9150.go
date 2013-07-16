package main

import (
	"bufio"
	"fmt"
	"github.com/tarm/goserial"
	"io"
	"time"
)

const (
	LSB_TO_RADIANS_PER_SECOND = (1.0 / 16.4) * 0.0174532925
	MPU_SLAVE_ADDR            = 0x68
	MPU_MAG_SLAVE_ADDR        = 0x0C
)

type MPU9150_IMUSensor struct {
	config     serial.Config
	port       io.ReadWriteCloser
	readBuffer *bufio.Reader
	enable     bool
}

func CreateMPU9150_IMUSensor(name string, baud int) *MPU9150_IMUSensor {
	cfg := serial.Config{Name: name, Baud: baud}
	return &MPU9150_IMUSensor{
		config: cfg,
		enable: true,
	}
}

func (self *MPU9150_IMUSensor) Connect() {
	if port, err := serial.OpenPort(&self.config); err != nil {
		panic("Failed to connect to MPU9150 IMU sensor.")
	} else {
		self.port = port
		self.readBuffer = bufio.NewReader(self.port)

		self.i2cMasterEnable()
		self.disableStreaming()
		self.wake()
	}
}

func (self *MPU9150_IMUSensor) Disconnect() {
	self.enable = false
}

func (self *MPU9150_IMUSensor) disableStreaming() {
	self.port.Write([]byte("SD\n"))
	self.readBuffer.ReadBytes('\n')
}

func (self *MPU9150_IMUSensor) wake() {
	self.writeRegister(0x68, 0x6b, 0x01)
	self.readBuffer.ReadBytes('\n') // drop the response
}

func (self *MPU9150_IMUSensor) readRegister(slaveAddr, regAddr, byteCount uint) []byte {
	self.port.Write([]byte(fmt.Sprintf("RD%02X%02X%02X\n", slaveAddr, regAddr, byteCount)))
	buf, _ := self.readBuffer.ReadBytes('\n')
	return buf
}

func (self *MPU9150_IMUSensor) writeRegister(slaveAddr, regAddr, newValue int) {
	self.port.Write([]byte(fmt.Sprintf("WR%02X%02X%02X\n", slaveAddr, regAddr, newValue)))
	self.readBuffer.ReadBytes('\n')
}

func (self *MPU9150_IMUSensor) sensorValueFromResponseAtIndex(response []byte, index uint) int16 {
	var byte_h, byte_l int16
	fmt.Sscanf(string(response[index:index+4]), "%02X%02X", &byte_h, &byte_l)
	return (byte_h << 8) + byte_l
}

func (self *MPU9150_IMUSensor) i2cMasterEnable() {
	self.writeRegister(MPU_SLAVE_ADDR, 0x6a, 1<<5) // i2c_mst_enable
	self.writeRegister(MPU_SLAVE_ADDR, 0x37, 0x00) // disable i2c input bypass
}

func (self *MPU9150_IMUSensor) i2cMasterDisable() {
	self.writeRegister(MPU_SLAVE_ADDR, 0x6a, 0x00) // i2c_mst_enable
	self.writeRegister(MPU_SLAVE_ADDR, 0x37, 0x02) // enable i2c input bypass
}

func (self *MPU9150_IMUSensor) Sample() *IMUSensorSample {

	// close the port if needed.
	if !self.enable {
		self.i2cMasterEnable()
		self.disableStreaming()
		self.port.Close()
		return nil
	}

	sampleTime := time.Now()

	/****** Read Accelerometer/Gyro ******/
	self.i2cMasterEnable()
	accelGyroRaw := self.readRegister(MPU_SLAVE_ADDR, 0x3b, 14)

	/****** Read Magnetometer ******/
	self.i2cMasterDisable()
	self.writeRegister(MPU_MAG_SLAVE_ADDR, 0x0a, 0x01) // request single sample

	// wait until magnetometer is ready
	for {
		ready := self.readRegister(MPU_MAG_SLAVE_ADDR, 0x02, 1)
		if ready[7] == '1' {
			break
		} else {
			time.Sleep(500)
		}
	}

	magRaw := self.readRegister(MPU_MAG_SLAVE_ADDR, 0x03, 6)

	return &IMUSensorSample{
		t: sampleTime,
		// accelerometer values
		ax: float32(self.sensorValueFromResponseAtIndex(accelGyroRaw, 6)),
		ay: float32(self.sensorValueFromResponseAtIndex(accelGyroRaw, 10)),
		az: float32(self.sensorValueFromResponseAtIndex(accelGyroRaw, 14)),
		// gyro values
		gx: float32(self.sensorValueFromResponseAtIndex(accelGyroRaw, 22)) * LSB_TO_RADIANS_PER_SECOND,
		gy: float32(self.sensorValueFromResponseAtIndex(accelGyroRaw, 26)) * LSB_TO_RADIANS_PER_SECOND,
		gz: float32(self.sensorValueFromResponseAtIndex(accelGyroRaw, 30)) * LSB_TO_RADIANS_PER_SECOND,
		// magnetometer values
		mx: float32(self.sensorValueFromResponseAtIndex(magRaw, 6)),
		my: float32(self.sensorValueFromResponseAtIndex(magRaw, 10)),
		mz: float32(self.sensorValueFromResponseAtIndex(magRaw, 14)),
	}
}
