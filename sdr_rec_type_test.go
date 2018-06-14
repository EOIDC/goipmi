/*
Copyright (c) 2014 EOITek, Inc. All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package ipmi

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSDRRecType_McDeviceLocator(t *testing.T) {

	deviceId := "AST2400"
	rec, err := NewSDRMcDeviceLocator(uint16(1), deviceId)
	assert.Nil(t, err)

	rec.sdrMcDeviceLocatorFields = sdrMcDeviceLocatorFields{
		DeviceSlaveAddr: 0x20,
		ChannelNumber:   0x00,
		PSNGI:           0x00,
		DeviceCap:       0xff,
		reserved:        [3]byte{0, 0, 0},
		EntityId:        0x00,
		EntityIns:       0x01,
		OEM:             0x00,
	}

	data, err := rec.MarshalBinary()
	assert.Nil(t, err)
	assert.NotNil(t, data)

	flen := 10 //sizeof sdrmcdevicelocatorfields
	assert.Equal(t, byte(0x01), data[0])
	assert.Equal(t, byte(0x00), data[1])
	assert.Equal(t, byte(SDR_RECORD_TYPE_MC_DEVICE_LOCATOR), data[3])
	assert.Equal(t, byte(flen+1+len(deviceId)), data[4])
	assert.Equal(t, byte(len(deviceId))|0xc0, data[flen+1+4])
	assert.Equal(t, deviceId, string(data[flen+1+4+1:]))
}

func TestSDRRecType_FullSensor(t *testing.T) {
	deviceId := "Ambient Temp"
	rec, err := NewSDRFullSensor(uint16(321), deviceId)
	assert.Nil(t, err)
	fmt.Println("DeviceId: ", rec.DeviceId())
	fmt.Println("RecordId: ", rec.RecordId())

	rec.sdrFullSensorFields = sdrFullSensorFields{
		SensorOwnerId:        0x20,
		SensorOwnerLUN:       0x00,
		SensorNumber:         0x01,
		EntityId:             0x37,
		EntityIns:            0x01,
		SensorInit:           0x5f,
		SensorCap:            0x68,
		SensorType:           SDRSensorType(0x01),
		ReadingType:          SDRSensorReadingType(0x01),
		AssertionEventMask:   0x3285,
		DeassertionEventMask: 0x3285,
		DiscreteReadingMask:  0x1b1b,
		Unit:                 0x80,
		BaseUnit:             0x01,
		ModifierUnit:         0x00,
		Linearization:        0x00,
		MTol:                 0x0201,
		Bacc:                 0x0200,
		Acc:                  0x30,
		RBexp:                0x00,
		AnalogFlag:           0x07,
		NominalReading:       0x97,
		NormalMax:            0xc5,
		NormalMin:            0x8b,
		SensorMax:            0x7f,
		SensorMin:            0x80,
		U_NR:                 0x29,
		U_C:                  0x27,
		U_NC:                 0x25,
		L_NR:                 0x00,
		L_C:                  0x03,
		L_NC:                 0x08,
		PositiveHysteresis:   0x02,
		NegativeHysteresis:   0x02,
		Reserved:             [2]byte{0, 0},
		OEM:                  0x00,
	}

	data, err := rec.MarshalBinary()
	assert.Nil(t, err)
	assert.NotNil(t, data)

	flen := 42
	assert.Equal(t, byte(0x41), data[0])
	assert.Equal(t, byte(0x01), data[1])
	assert.Equal(t, byte(SDR_RECORD_TYPE_FULL_SENSOR), data[3])
	assert.Equal(t, byte(flen+1+len(deviceId)), data[4])
	assert.Equal(t, byte(len(deviceId))|0xc0, data[flen+1+4])
	assert.Equal(t, deviceId, string(data[flen+1+4+1:]))

	//test M,B
	assert.Equal(t, byte(0x01), data[24])
	assert.Equal(t, byte(0x02), data[25])
	assert.Equal(t, byte(0x00), data[26])
	assert.Equal(t, byte(0x02), data[27])
	assert.Equal(t, byte(0x30), data[28])
	assert.Equal(t, byte(0x00), data[29])

}

func TestSimulatorSDR_L_SetMBExp(t *testing.T) {
	r2, _ := NewSDRFullSensor(2, "System 3.3V")
	r2.SetMBExp(2, 0, 0, 0)
	assert.Equal(t, uint16(0x0002), r2.MTol)
	assert.Equal(t, uint16(0x0000), r2.Bacc)
	r2.SetMBExp(-2, 5, 0, -2)
	assert.Equal(t, uint16(0xc0fe), r2.MTol)
	assert.Equal(t, uint16(0x0005), r2.Bacc)
	assert.Equal(t, uint8(0xe0), r2.RBexp)
	r2.SetMBExp(-2, -5, -2, -2)
	assert.Equal(t, uint16(0xc0fe), r2.MTol)
	assert.Equal(t, uint16(0xc0fb), r2.Bacc)
	assert.Equal(t, uint8(0xee), r2.RBexp)
}

func TestSimulatorSDR_L_GetMBExp(t *testing.T) {
	r1, _ := NewSDRFullSensor(2, "System 3.3V")
	r1.SetMBExp(2, 0, 0, 0)
	m1, b1, be1, re1 := r1.GetMBExp()
	assert.Equal(t, int16(2), m1)
	assert.Equal(t, int16(0), b1)
	assert.Equal(t, int8(0), be1)
	assert.Equal(t, int8(0), re1)

	r1.SetMBExp(-2, 5, 0, -2)
	m2, b2, be2, re2 := r1.GetMBExp()
	assert.Equal(t, int16(-2), m2)
	assert.Equal(t, int16(5), b2)
	assert.Equal(t, int8(0), be2)
	assert.Equal(t, int8(-2), re2)

	r1.SetMBExp(-2, -5, -2, -2)
	m3, b3, be3, re3 := r1.GetMBExp()
	assert.Equal(t, int16(-2), m3)
	assert.Equal(t, int16(-5), b3)
	assert.Equal(t, int8(-2), be3)
	assert.Equal(t, int8(-2), re3)
}

func TestGetMBExp(t *testing.T) {
	r, _ := NewSDRFullSensor(2, "Inlet Temp")
	r.MTol = 0x0064
	r.Bacc = 0x0000
	r.Acc = 0x00
	r.RBexp = 0xe0
	m,b,be,re := r.GetMBExp()
	assert.Equal(t, int16(100), m)
	assert.Equal(t, int16(0), b)
	assert.Equal(t, int8(0), be)
	assert.Equal(t, int8(-2), re)
}

func TestSimulatorSDR_L_CalValue1(t *testing.T) {
	r, _ := NewSDRFullSensor(2, "System 3.3V")
	//r.Unit = 0x80 //2's complement signed
	r.Unit = 0x00          //unsigned
	r.BaseUnit = 0x04      //Voltage
	r.Linearization = 0x00 //no linearization
	r.SetMBExp(2, 0, 0, -2)
	v := r.CalValue(3.36)
	assert.Equal(t, uint8(0xa8), v)
}

func TestSimulatorSDR_L_CalValue2(t *testing.T) {
	r, _ := NewSDRFullSensor(2, "Exhaust Temp")
	r.Unit = 0x80          //2's complement signed
	r.BaseUnit = 0x01      //Temprature
	r.Linearization = 0x00 //no linearization
	r.SetMBExp(1, 0, 0, 0)
	v := r.CalValue(23.0)
	assert.Equal(t, uint8(0x17), v)
}

func TestSimulatorSDR_L_CalValue3(t *testing.T) {
	r, _ := NewSDRFullSensor(2, "CPU1 DTS")
	r.Unit = 0x80          //2's complement signed
	r.BaseUnit = 0x01      //Temprature
	r.Linearization = 0x00 //no linearization
	r.SetMBExp(1, 0, 0, 0)
	v := r.CalValue(-49.0)
	assert.Equal(t, uint8(0xcf), v)
}

func TestSimulatorSDR_L_CalValue4(t *testing.T) {
	r, _ := NewSDRFullSensor(2, "Fan 1")
	r.Unit = 0x00          //2's complement signed
	r.BaseUnit = 0x12      //RPM
	r.Linearization = 0x00 //no linearization
	r.SetMBExp(63, 0, 0, 0)
	v := r.CalValue(2583.0)
	assert.Equal(t, uint8(0x29), v)
}

func TestSimulatorSDR_L_FullSensorMarshaling(t *testing.T) {
	r1, _ := NewSDRFullSensor(2, "System 3.3V")
	//mock some feilds and used to assert
	r1.SensorOwnerId = 0x01
	r1.Unit = 0x80
	r1.BaseUnit = 0x04
	r1.Linearization = 0x02
	r1.SensorMax = 0x03

	bin, err := r1.MarshalBinary()
	assert.Nil(t, err)

	r2, _ := NewSDRFullSensor(0, "")
	err = r2.UnmarshalBinary(bin)
	assert.Nil(t, err)

	assert.Equal(t, r2.SensorOwnerId, r1.SensorOwnerId)
	assert.Equal(t, r2.Unit, r1.Unit)
	assert.Equal(t, r2.BaseUnit, r1.BaseUnit)
	assert.Equal(t, r2.Linearization, r1.Linearization)
	assert.Equal(t, r2.SensorMax, r1.SensorMax)
	assert.Equal(t, r2.DeviceId(), r1.DeviceId())
	assert.Equal(t, r2.RecordId(), r1.RecordId())
	assert.Equal(t, r2.RecordType(), r1.RecordType())

}
