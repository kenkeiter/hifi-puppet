/*
sensor.go by Kenneth Keiter <ken@kenkeiter.com>
Copyright (c) 2013  High Fidelity, Inc. All rights reserved.

Basic sensor constructs.
*/

package main

import (
	"fmt"
	"time"
)

type IMUSensorSample struct {
	t          time.Time
	gx, gy, gz float32 // gyro measurement
	ax, ay, az float32 // accelerometer measurement
	mx, my, mz float32 // magnetometer measurement
}

func (self *IMUSensorSample) Clone() *IMUSensorSample {
	return &IMUSensorSample{
		t:  self.t,
		ax: self.ax, ay: self.ay, az: self.az,
		gx: self.gx, gy: self.gy, gz: self.gz,
		mx: self.mx, my: self.my, mz: self.mz,
	}
}

func (self *IMUSensorSample) Inspect() string {
	return fmt.Sprintf("<Sample [ax:%.2f ay:%.2f az:%.2f] [gx:%.2f gy:%.2f gz:%.2f] [mx:%.2f my:%.2f mz:%.2f]>",
		self.ax, self.ay, self.az, self.gx, self.gy, self.gz, self.mx, self.my, self.mz)
}
