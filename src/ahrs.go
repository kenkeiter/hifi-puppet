/*
 * Implements a quaternion-based AHRS filter.
 */

package main

import (
	"fmt"
	"math"
	"time"
)

type AHRSQuaternionFrame struct {
	sample     *IMUSensorSample
	x, y, z, w float32
}

func (self *AHRSQuaternionFrame) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("{\"quaternion\": [%f, %f, %f, %f]}",
		self.x, self.y, self.z, self.w)), nil
}

type AHRSFilter struct {
	sampleFreq        float32
	beta              float32
	q0, q1, q2, q3    float32
	sensor            *MPU9150_IMUSensor
	lastFrame         *AHRSQuaternionFrame
	motionSubscribers []*WebsocketClient
}

func NewAHRSFilter(sensor *MPU9150_IMUSensor, betaDef, sampleFreq float32) *AHRSFilter {
	filter := &AHRSFilter{
		sensor:            sensor,
		beta:              betaDef,
		q0:                1.0,
		q1:                0.0,
		q2:                0.0,
		q3:                0.0,
		sampleFreq:        sampleFreq,
		motionSubscribers: make([]*WebsocketClient, 0),
	}
	go filter.beginSampling()
	return filter
}

func (self *AHRSFilter) beginSampling() {
	for {
		self.Update(self.sensor.Sample())
		for _, subscriber := range self.motionSubscribers {
			subscriber.Motion(self.lastFrame)
		}
		time.Sleep(time.Duration(time.Second / time.Duration(self.sampleFreq)))
	}
}

func (self *AHRSFilter) UpdateSampleFrequency(newFreq float32) {
	self.sampleFreq = newFreq
}

func (self *AHRSFilter) Reset() {
	self.q0 = 1.0
	self.q1 = 0.0
	self.q2 = 0.0
	self.q3 = 0.0
}

func (self *AHRSFilter) Subscribe(s *WebsocketClient) {
	self.motionSubscribers = append(self.motionSubscribers, s)
}

func (self *AHRSFilter) Unsubscribe(s *WebsocketClient) {
	for i, test_sub := range self.motionSubscribers {
		if test_sub == s {
			self.motionSubscribers[i] = self.motionSubscribers[len(self.motionSubscribers)-1]
			self.motionSubscribers = self.motionSubscribers[:len(self.motionSubscribers)-1]
		}
	}
}

func (self *AHRSFilter) Inspect() string {
	return fmt.Sprintf("<Quaternion [x:%.2f y:%.2f z:%.2f w:%.2f]>",
		self.q0, self.q1, self.q2, self.q3)
}

func (self *AHRSFilter) LastQuaternionFrame() *AHRSQuaternionFrame {
	frame := self.lastFrame
	return frame
}

func (self *AHRSFilter) Update(originalSample *IMUSensorSample) {
	sample := originalSample.Clone()

	var recipNorm float32
	var s0, s1, s2, s3 float32
	var qDot1, qDot2, qDot3, qDot4 float32
	var hx, hy float32
	var _2q0mx, _2q0my, _2q0mz, _2q1mx, _2bx, _2bz, _4bx, _4bz,
		_2q0, _2q1, _2q2, _2q3, _2q0q2, _2q2q3, q0q0, q0q1,
		q0q2, q0q3, q1q1, q1q2, q1q3, q2q2, q2q3, q3q3 float32

	// Rate of change of quaternion from gyroscope.
	qDot1 = 0.5 * (-self.q1*sample.gx - self.q2*sample.gy - self.q3*sample.gz)
	qDot2 = 0.5 * (self.q0*sample.gx + self.q2*sample.gz - self.q3*sample.gy)
	qDot3 = 0.5 * (self.q0*sample.gy - self.q1*sample.gz + self.q3*sample.gx)
	qDot4 = 0.5 * (self.q0*sample.gz + self.q1*sample.gy - self.q2*sample.gx)

	// Compute feedback only if accelerometer measurement is valid (avoids NaN)
	if !(sample.ax == 0.0 && sample.ay == 0.0 && sample.az == 0.0) {

		/* Normalize accelerometer */
		recipNorm = float32(1.0 / math.Sqrt(float64(sample.ax*sample.ax+sample.ay*sample.ay+sample.az*sample.az)))
		sample.ax *= recipNorm
		sample.ay *= recipNorm
		sample.az *= recipNorm

		/* Normalize magnetometer */
		recipNorm = float32(1.0 / math.Sqrt(float64(sample.mx*sample.mx+sample.my*sample.my+sample.mz*sample.mz)))
		sample.mx *= recipNorm
		sample.my *= recipNorm
		sample.mz *= recipNorm

		/* Avoid repeated maths. */
		_2q0mx = 2.0 * self.q0 * sample.mx
		_2q0my = 2.0 * self.q0 * sample.my
		_2q0mz = 2.0 * self.q0 * sample.mz
		_2q1mx = 2.0 * self.q1 * sample.mx
		_2q0 = 2.0 * self.q0
		_2q1 = 2.0 * self.q1
		_2q2 = 2.0 * self.q2
		_2q3 = 2.0 * self.q3
		_2q0q2 = 2.0 * self.q0 * self.q2
		_2q2q3 = 2.0 * self.q2 * self.q3
		q0q0 = self.q0 * self.q0
		q0q1 = self.q0 * self.q1
		q0q2 = self.q0 * self.q2
		q0q3 = self.q0 * self.q3
		q1q1 = self.q1 * self.q1
		q1q2 = self.q1 * self.q2
		q1q3 = self.q1 * self.q3
		q2q2 = self.q2 * self.q2
		q2q3 = self.q2 * self.q3
		q3q3 = self.q3 * self.q3

		/* Reference direction of Earth's magnetic field */
		hx = sample.mx*q0q0 - _2q0my*self.q3 + _2q0mz*self.q2 + sample.mx*q1q1 +
			_2q1*sample.my*self.q2 + _2q1*sample.mz*self.q3 - sample.mx*q2q2 - sample.mx*q3q3
		hy = _2q0mx*self.q3 + sample.my*q0q0 - _2q0mz*self.q1 + _2q1mx*self.q2 -
			sample.my*q1q1 + sample.my*q2q2 + _2q2*sample.mz*self.q3 - sample.my*q3q3
		_2bx = float32(math.Sqrt(float64(hx*hx + hy*hy)))
		_2bz = -_2q0mx*self.q2 + _2q0my*self.q1 + sample.mz*q0q0 + _2q1mx*self.q3 -
			sample.mz*q1q1 + _2q2*sample.my*self.q3 - sample.mz*q2q2 + sample.mz*q3q3
		_4bx = 2.0 * _2bx
		_4bz = 2.0 * _2bz

		/* Gradient decent algorithm corrective step */
		s0 = -_2q2*(2.0*q1q3-_2q0q2-sample.ax) + _2q1*(2.0*q0q1+_2q2q3-sample.ay) -
			_2bz*self.q2*(_2bx*(0.5-q2q2-q3q3)+_2bz*(q1q3-q0q2)-sample.mx) +
			(-_2bx*self.q3+_2bz*self.q1)*(_2bx*(q1q2-q0q3)+_2bz*(q0q1+q2q3)-sample.my) +
			_2bx*self.q2*(_2bx*(q0q2+q1q3)+_2bz*(0.5-q1q1-q2q2)-sample.mz)
		s1 = _2q3*(2.0*q1q3-_2q0q2-sample.ax) + _2q0*(2.0*q0q1+_2q2q3-sample.ay) -
			4.0*self.q1*(1-2.0*q1q1-2.0*q2q2-sample.az) + _2bz*self.q3*(_2bx*(0.5-q2q2-q3q3)+
			_2bz*(q1q3-q0q2)-sample.mx) + (_2bx*self.q2+_2bz*self.q0)*(_2bx*(q1q2-q0q3)+_2bz*(q0q1+q2q3)-sample.my) +
			(_2bx*self.q3-_4bz*self.q1)*(_2bx*(q0q2+q1q3)+_2bz*(0.5-q1q1-q2q2)-sample.mz)
		s2 = -_2q0*(2.0*q1q3-_2q0q2-sample.ax) + _2q3*(2.0*q0q1+_2q2q3-sample.ay) -
			4.0*self.q2*(1-2.0*q1q1-2.0*q2q2-sample.az) + (-_4bx*self.q2-_2bz*self.q0)*(_2bx*(0.5-q2q2-q3q3)+
			_2bz*(q1q3-q0q2)-sample.mx) + (_2bx*self.q1+_2bz*self.q3)*(_2bx*(q1q2-q0q3)+_2bz*(q0q1+q2q3)-sample.my) +
			(_2bx*self.q0-_4bz*self.q2)*(_2bx*(q0q2+q1q3)+_2bz*(0.5-q1q1-q2q2)-sample.mz)
		s3 = _2q1*(2.0*q1q3-_2q0q2-sample.ax) + _2q2*(2.0*q0q1+_2q2q3-sample.ay) +
			(-_4bx*self.q3+_2bz*self.q1)*(_2bx*(0.5-q2q2-q3q3)+_2bz*(q1q3-q0q2)-sample.mx) +
			(-_2bx*self.q0+_2bz*self.q2)*(_2bx*(q1q2-q0q3)+_2bz*(q0q1+q2q3)-sample.my) +
			_2bx*self.q1*(_2bx*(q0q2+q1q3)+_2bz*(0.5-q1q1-q2q2)-sample.mz)

		recipNorm = float32(1.0 / math.Sqrt(float64(s0*s0+s1*s1+s2*s2+s3*s3))) // normalize step magnitude
		s0 *= recipNorm
		s1 *= recipNorm
		s2 *= recipNorm
		s3 *= recipNorm

		// Apply feedback step
		qDot1 -= self.beta * s0
		qDot2 -= self.beta * s1
		qDot3 -= self.beta * s2
		qDot4 -= self.beta * s3
	}

	// Integrate rate of change of quaternion to yield quaternion
	self.q0 += qDot1 * (1.0 / self.sampleFreq)
	self.q1 += qDot2 * (1.0 / self.sampleFreq)
	self.q2 += qDot3 * (1.0 / self.sampleFreq)
	self.q3 += qDot4 * (1.0 / self.sampleFreq)

	// Normalize quaternion
	recipNorm = float32(1.0 / math.Sqrt(float64(self.q0*self.q0+self.q1*self.q1+self.q2*self.q2+self.q3*self.q3)))
	self.q0 *= recipNorm
	self.q1 *= recipNorm
	self.q2 *= recipNorm
	self.q3 *= recipNorm

	self.lastFrame = &AHRSQuaternionFrame{
		sample: originalSample,
		x:      self.q0,
		y:      self.q1,
		z:      self.q2,
		w:      self.q3,
	}
}
