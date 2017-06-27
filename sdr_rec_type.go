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
	"bytes"
	"encoding/binary"
	"errors"
	//"fmt"
	"math"
)

var (
	ErrDeviceIdMustLess16 = errors.New("Device Id must be less or equal to 16 bytes length")
	ErrUnitNotSupport     = errors.New("Unit not support, only support unsigned and 2's complement signed")
	ErrMZero              = errors.New("M mustn't be 0")
)

type SDRRecord interface {
	DeviceId() string
	RecordId() uint16
	RecordType() SDRRecordType
}

type SDRRecordHeader struct {
	recordId   uint16
	SDRVersion uint8
	rtype      SDRRecordType
}

// section 43.9
type sdrMcDeviceLocatorFields struct { //size 10
	DeviceSlaveAddr uint8
	ChannelNumber   uint8

	PSNGI     uint8
	DeviceCap uint8
	reserved  [3]byte
	EntityId  uint8
	EntityIns uint8
	OEM       uint8
}

type SDRMcDeviceLocator struct {
	SDRRecordHeader
	sdrMcDeviceLocatorFields
	deviceId string
}

func NewSDRMcDeviceLocator(id uint16, name string) (*SDRMcDeviceLocator, error) {
	if len(name) > 16 {
		return nil, ErrDeviceIdMustLess16
	}
	r := &SDRMcDeviceLocator{}
	r.recordId = id
	r.rtype = SDR_RECORD_TYPE_MC_DEVICE_LOCATOR
	r.SDRVersion = 0x51
	r.deviceId = name
	return r, nil
}

func (r *SDRMcDeviceLocator) DeviceId() string {
	return r.deviceId
}

func (r *SDRMcDeviceLocator) RecordId() uint16 {
	return r.recordId
}

func (r *SDRMcDeviceLocator) RecordType() SDRRecordType {
	return r.rtype
}

func (r *SDRMcDeviceLocator) MarshalBinary() (data []byte, err error) {
	hb := new(bytes.Buffer)
	fb := new(bytes.Buffer)
	db := new(bytes.Buffer)
	binary.Write(hb, binary.LittleEndian, r.SDRRecordHeader)
	binary.Write(fb, binary.LittleEndian, r.sdrMcDeviceLocatorFields)
	db.WriteByte(byte(len(r.DeviceId())))
	db.WriteString(r.DeviceId())

	//merge all
	recLen := uint8(fb.Len() + db.Len())
	hb.WriteByte(byte(recLen))
	hb.Write(fb.Bytes())
	hb.Write(db.Bytes())
	return hb.Bytes(), nil
}

// section 43.1
type sdrFullSensorFields struct { //size 42
	SensorOwnerId        uint8
	SensorOwnerLUN       uint8
	SensorNumber         uint8
	EntityId             uint8
	EntityIns            uint8
	SensorInit           uint8
	SensorCap            uint8
	SensorType           SDRSensorType
	ReadingType          SDRSensorReadingType
	AssertionEventMask   uint16
	DeassertionEventMask uint16
	DiscreteReadingMask  uint16
	Unit                 uint8
	BaseUnit             uint8
	ModifierUnit         uint8
	Linearization        uint8
	MTol                 uint16
	Bacc                 uint16
	Acc                  uint8
	RBexp                uint8
	AnalogFlag           uint8
	NominalReading       uint8
	NormalMax            uint8
	NormalMin            uint8
	SensorMax            uint8
	SensorMin            uint8
	U_NR                 uint8
	U_C                  uint8
	U_NC                 uint8
	L_NR                 uint8
	L_C                  uint8
	L_NC                 uint8
	PositiveHysteresis   uint8
	NegativeHysteresis   uint8
	reserved             [2]byte
	OEM                  uint8
}

type SDRFullSensor struct {
	SDRRecordHeader
	sdrFullSensorFields
	deviceId string
}

func NewSDRFullSensor(id uint16, name string) (*SDRFullSensor, error) {
	if len(name) > 16 {
		return nil, ErrDeviceIdMustLess16
	}
	r := &SDRFullSensor{}
	r.recordId = id
	r.rtype = SDR_RECORD_TYPE_FULL_SENSOR
	r.SDRVersion = 0x51
	r.deviceId = name
	return r, nil
}

func (r *SDRFullSensor) DeviceId() string {
	return r.deviceId
}

func (r *SDRFullSensor) RecordId() uint16 {
	return r.recordId
}

func (r *SDRFullSensor) RecordType() SDRRecordType {
	return r.rtype
}

//M: 10bit signed 2's complement
//B: 10bit signed 2's complement
//Bexp: 4bit signed 2's complement
//Rexp: 4bit signed 2's complement
func (r *SDRFullSensor) SetMBExp(M int16, B int16, Bexp int8, Rexp int8) {

	r.MTol = 0
	r.Bacc = 0
	r.RBexp = 0

	_M := uint16(math.Abs(float64(M)))
	_M = _M & 0x01ff //mask leave low 9bit
	if M < 0 {
		_M = (((^_M) + 1) & 0x01ff) | 0x0200
	}
	r.MTol = r.MTol | (_M & 0x00ff)
	r.MTol = r.MTol | ((_M << 6) & 0xc000)

	_B := uint16(math.Abs(float64(B)))
	_B = _B & 0x01ff //mask leave low 9bit
	if B < 0 {
		_B = (((^_B) + 1) & 0x01ff) | 0x0200
	}
	r.Bacc = r.Bacc | (_B & 0x00ff)
	r.Bacc = r.Bacc | ((_B << 6) & 0xc000)

	_Bexp := uint8(math.Abs(float64(Bexp)))
	_Bexp = _Bexp & 0x07 //mask leeve low 3bit
	if Bexp < 0 {
		_Bexp = (((^_Bexp) + 1) & 0x07) | 0x08
	}
	r.RBexp = r.RBexp | (_Bexp & 0x0f)

	_Rexp := uint8(math.Abs(float64(Rexp)))
	_Rexp = _Rexp & 0x07 //mask leave low 3bit
	if Rexp < 0 {
		_Rexp = (((^_Rexp) + 1) & 0x07) | 0x08
	}
	r.RBexp = r.RBexp | ((_Rexp << 4) & 0xf0)

}

func (r *SDRFullSensor) GetMBExp() (M int16, B int16, Bexp int8, Rexp int8) {
	_M := uint16(((r.MTol & 0xc000) >> 6) | (r.MTol & 0x00ff))
	if (_M & 0x0200) == 0x0200 { //most significate is 1, mean signed
		//fmt.Printf("%d,0x%x\n", int16((_M & 0xfdff)), (_M & 0xfdff))
		M = int16((_M & 0xfdff)) - 512 //2^9
	} else {
		M = int16(_M & 0xfdff)
	}

	_B := uint16(((r.Bacc & 0xc000) >> 6) | (r.Bacc & 0x00ff))
	if (_B & 0x0200) == 0x0200 { //most significate is 1, mean signed
		B = int16((_B & 0xfdff)) - 512 //2^9
	} else {
		B = int16(_B & 0xfdff)
	}

	_Bexp := uint8(r.RBexp & 0x0f)
	if (_Bexp & 0x08) == 0x08 {
		Bexp = int8((_Bexp & 0xf7)) - 8 //2^3
	} else {
		Bexp = int8(_Bexp & 0xf7)
	}

	_Rexp := uint8((r.RBexp & 0xf0) >> 4)
	if (_Rexp & 0x08) == 0x08 {
		Rexp = int8((_Rexp & 0xf7)) - 8 //2^3
	} else {
		Rexp = int8(_Rexp & 0xf7)
	}

	return
}

// calculate the given value into the SDR reading value, using current M,B,Bexp,Rexp setting
func (r *SDRFullSensor) CalValue(value float64) uint8 {
	M, B, Bexp, Rexp := r.GetMBExp()
	if M == 0 {
		panic(ErrMZero)
	}

	//y=(M x V + B x pow(10,Bexp)) x pow(10,Rexp)
	//know y, cal V
	var neg bool = false
	v := (value/math.Pow(10, float64(Rexp)) - float64(B)*math.Pow(10, float64(Bexp))) / float64(M)
	if v < 0 {
		neg = true
	}
	v = math.Abs(v)
	uv := uint8(v)
	if neg {
		if (r.Unit & 0xc0) == 0x80 {
			return ((128 - uv) | 0x80)
		} else {
			panic(ErrUnitNotSupport)
		}
	} else {
		if (r.Unit & 0xc0) == 0x00 {
			return uv
		} else if (r.Unit & 0xc0) == 0x80 {
			return uv & 0x7f
		} else {
			panic(ErrUnitNotSupport)
		}
	}
}

func (r *SDRFullSensor) MarshalBinary() (data []byte, err error) {
	hb := new(bytes.Buffer)
	fb := new(bytes.Buffer)
	db := new(bytes.Buffer)
	binary.Write(hb, binary.LittleEndian, r.SDRRecordHeader)
	binary.Write(fb, binary.LittleEndian, r.sdrFullSensorFields)
	db.WriteByte(byte(len(r.DeviceId())))
	db.WriteString(r.DeviceId())

	//merge all
	recLen := uint8(fb.Len() + db.Len())
	hb.WriteByte(byte(recLen))
	hb.Write(fb.Bytes())
	hb.Write(db.Bytes())
	return hb.Bytes(), nil
}

type SDRCompactSensor struct {
}
