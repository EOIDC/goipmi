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
	ErrDeviceIdMustLess16   = errors.New("Device Id must be less or equal to 16 bytes length")
	ErrUnitNotSupport       = errors.New("Unit not support, only support unsigned and 2's complement signed")
	ErrMZero                = errors.New("M mustn't be 0")
	ErrIdStringLenNotMatch  = errors.New("Length of the Id string is mismatch")
	ErrSensorReadUnavail    = errors.New("Sensor Reading Unavailable")
	ErrNotFoundTheSensorNum = errors.New("failed to found the SensorNumber")
)

var sdrRecordValueBasicUnit []string = []string{
	"unspecified",
	"degrees C", "degrees F", "degrees K",
	"Volts", "Amps", "Watts", "Joules",
	"Coulombs", "VA", "Nits",
	"lumen", "lux", "Candela",
	"kPa", "PSI", "Newton",
	"CFM", "RPM", "Hz",
	"microsecond", "millisecond", "second", "minute", "hour",
	"day", "week", "mil", "inches", "feet", "cu in", "cu feet",
	"mm", "cm", "m", "cu cm", "cu m", "liters", "fluid ounce",
	"radians", "steradians", "revolutions", "cycles",
	"gravities", "ounce", "pound", "ft-lb", "oz-in", "gauss",
	"gilberts", "henry", "millihenry", "farad", "microfarad",
	"ohms", "siemens", "mole", "becquerel", "PPM", "reserved",
	"Decibels", "DbA", "DbC", "gray", "sievert",
	"color temp deg K", "bit", "kilobit", "megabit", "gigabit",
	"byte", "kilobyte", "megabyte", "gigabyte", "word", "dword",
	"qword", "line", "hit", "miss", "retry", "reset",
	"overflow", "underrun", "collision", "packets", "messages",
	"characters", "error", "correctable error", "uncorrectable error"}
var sdrRecordValueSensorType []string = []string{
	"reserved",
	"Temperature", "Voltage", "Current", "Fan",
	"Physical Security", "Platform Security", "Processor",
	"Power Supply", "Power Unit", "Cooling Device", "Other",
	"Memory", "Drive Slot / Bay", "POST Memory Resize",
	"System Firmwares", "Event Logging Disabled", "Watchdog1",
	"System Event", "Critical Interrupt", "Button",
	"Module / Board", "Microcontroller", "Add-in Card",
	"Chassis", "Chip Set", "Other FRU", "Cable / Interconnect",
	"Terminator", "System Boot Initiated", "Boot Error",
	"OS Boot", "OS Critical Stop", "Slot / Connector",
	"System ACPI Power State", "Watchdog2", "Platform Alert",
	"Entity Presence", "Monitor ASIC", "LAN",
	"Management Subsys Health", "Battery", "Session Audit",
	"Version Change", "FRU State"}

type SDRRecord interface {
	DeviceId() string
	RecordId() uint16
	RecordType() SDRRecordType
}

type SDRRecordHeader struct {
	Recordid   uint16
	SDRVersion uint8
	Rtype      SDRRecordType
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
	Deviceid string
}

func NewSDRMcDeviceLocator(id uint16, name string) (*SDRMcDeviceLocator, error) {
	if len(name) > 16 {
		return nil, ErrDeviceIdMustLess16
	}
	r := &SDRMcDeviceLocator{}
	r.Recordid = id
	r.Rtype = SDR_RECORD_TYPE_MC_DEVICE_LOCATOR
	r.SDRVersion = 0x51
	r.Deviceid = name
	return r, nil
}

func (r *SDRMcDeviceLocator) DeviceId() string {
	return r.Deviceid
}

func (r *SDRMcDeviceLocator) RecordId() uint16 {
	return r.Recordid
}

func (r *SDRMcDeviceLocator) RecordType() SDRRecordType {
	return r.Rtype
}

func (r *SDRMcDeviceLocator) MarshalBinary() (data []byte, err error) {
	hb := new(bytes.Buffer)
	fb := new(bytes.Buffer)
	db := new(bytes.Buffer)
	binary.Write(hb, binary.LittleEndian, r.SDRRecordHeader)
	binary.Write(fb, binary.LittleEndian, r.sdrMcDeviceLocatorFields)
	idl := generateIdLen(uint8(len(r.DeviceId())))
	db.WriteByte(idl)
	db.WriteString(r.DeviceId())

	//merge all
	recLen := uint8(fb.Len() + db.Len())
	hb.WriteByte(byte(recLen))
	hb.Write(fb.Bytes())
	hb.Write(db.Bytes())
	return hb.Bytes(), nil
}

// section 43.9
type sdrFruDeviceLocatorFields struct { //size 10
	DeviceAccAddr     uint8
	FRUDeviceID       uint8
	LogPhyAccLUNBusID uint8
	ChannNum          uint8
	reserved          uint8
	DeviceType        uint8
	DevTypeModif      uint8
	FruEntityId       uint8
	FruEntityInst     uint8
	Oem               uint8
}

type SDRFruDeviceLocator struct {
	SDRRecordHeader
	sdrFruDeviceLocatorFields
	Deviceid string
}

func NewSDRFruDeviceLocator(id uint16, name string) (*SDRFruDeviceLocator, error) {
	if len(name) > 16 {
		return nil, ErrDeviceIdMustLess16
	}
	r := &SDRFruDeviceLocator{}
	r.Recordid = id
	r.Rtype = SDR_RECORD_TYPE_FRU_DEVICE_LOCATOR
	r.SDRVersion = 0x51
	r.Deviceid = name
	return r, nil
}

func (r *SDRFruDeviceLocator) DeviceId() string {
	return r.Deviceid
}

func (r *SDRFruDeviceLocator) RecordId() uint16 {
	return r.Recordid
}

func (r *SDRFruDeviceLocator) RecordType() SDRRecordType {
	return r.Rtype
}

func (r *SDRFruDeviceLocator) MarshalBinary() (data []byte, err error) {
	hb := new(bytes.Buffer)
	fb := new(bytes.Buffer)
	db := new(bytes.Buffer)
	binary.Write(hb, binary.LittleEndian, r.SDRRecordHeader)
	binary.Write(fb, binary.LittleEndian, r.sdrFruDeviceLocatorFields)
	idl := generateIdLen(uint8(len(r.DeviceId())))
	db.WriteByte(idl)
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
	Reserved             [2]byte
	OEM                  uint8
}

type SDRFullSensor struct {
	SDRRecordHeader
	sdrFullSensorFields
	Deviceid string
}

func NewSDRFullSensor(id uint16, name string) (*SDRFullSensor, error) {
	if len(name) > 16 {
		return nil, ErrDeviceIdMustLess16
	}
	r := &SDRFullSensor{}
	r.Recordid = id
	r.Rtype = SDR_RECORD_TYPE_FULL_SENSOR
	r.SDRVersion = 0x51
	r.Deviceid = name
	return r, nil
}

func (r *SDRFullSensor) DeviceId() string {
	return r.Deviceid
}

func (r *SDRFullSensor) RecordId() uint16 {
	return r.Recordid
}

func (r *SDRFullSensor) RecordType() SDRRecordType {
	return r.Rtype
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

//parse id string len, get type and actual len
// section 43.15
func parseIdLen(len uint8) uint8 {
	return len & 0x0f
}

func generateIdLen(len uint8) uint8 {
	return len | 0xc0
}

func (r *SDRFullSensor) MarshalBinary() (data []byte, err error) {
	hb := new(bytes.Buffer)
	fb := new(bytes.Buffer)
	db := new(bytes.Buffer)
	binary.Write(hb, binary.LittleEndian, r.SDRRecordHeader)
	binary.Write(fb, binary.LittleEndian, r.sdrFullSensorFields)
	idl := generateIdLen(uint8(len(r.DeviceId())))
	db.WriteByte(idl)
	db.WriteString(r.DeviceId())

	//merge all
	recLen := uint8(fb.Len() + db.Len())
	hb.WriteByte(byte(recLen))
	hb.Write(fb.Bytes())
	hb.Write(db.Bytes())
	return hb.Bytes(), nil
}

func (r *SDRFullSensor) UnmarshalBinary(data []byte) error {
	buffer := bytes.NewReader(data)
	err := binary.Read(buffer, binary.LittleEndian, &r.SDRRecordHeader)
	if err != nil {
		return err
	}

	//skip the record length
	_, err = buffer.ReadByte()
	if err != nil {
		return err
	}

	binary.Read(buffer, binary.LittleEndian, &r.sdrFullSensorFields)

	idLen, err := buffer.ReadByte()
	if err != nil {
		return err
	}

	idLen = parseIdLen(idLen)

	id := make([]byte, int(idLen))
	n, err := buffer.Read(id)
	if err != nil || n != int(idLen) {
		return ErrIdStringLenNotMatch
	}

	r.Deviceid = string(id)

	return nil
}

// section 43.2
type sdrCompactSensorFields struct { //size 26
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
	SensorRecSharing     uint16
	PThresHysteresisVal  uint8
	NThresHysteresisVal  uint8
	Reserved             [2]byte
	OEM                  uint8
	IDStringTypeLen      uint8
}
type SDRCompactSensor struct {
	SDRRecordHeader
	sdrCompactSensorFields
	Deviceid string
}

func NewSDRCompactSensor(id uint16, name string) (*SDRCompactSensor, error) {
	if len(name) > 16 {
		return nil, ErrDeviceIdMustLess16
	}
	r := &SDRCompactSensor{}
	r.Recordid = id
	r.Rtype = SDR_RECORD_TYPE_COMPACT_SENSOR
	r.SDRVersion = 0x51
	r.Deviceid = name
	return r, nil
}

func (r *SDRCompactSensor) MarshalBinary() (data []byte, err error) {
	hb := new(bytes.Buffer)
	fb := new(bytes.Buffer)
	db := new(bytes.Buffer)
	binary.Write(hb, binary.LittleEndian, r.SDRRecordHeader)
	binary.Write(fb, binary.LittleEndian, r.sdrCompactSensorFields)
	idl := generateIdLen(uint8(len(r.DeviceId())))
	db.WriteByte(idl)
	db.WriteString(r.DeviceId())

	//merge all
	recLen := uint8(fb.Len() + db.Len())
	hb.WriteByte(byte(recLen))
	hb.Write(fb.Bytes())
	hb.Write(db.Bytes())
	return hb.Bytes(), nil
}
func (r *SDRCompactSensor) DeviceId() string {
	return r.Deviceid
}

func (r *SDRCompactSensor) RecordId() uint16 {
	return r.Recordid
}

func (r *SDRCompactSensor) RecordType() SDRRecordType {
	return r.Rtype
}
func (r *SDRCompactSensor) UnmarshalBinary(data []byte) error {
	buffer := bytes.NewReader(data)
	err := binary.Read(buffer, binary.LittleEndian, &r.SDRRecordHeader)
	if err != nil {
		return err
	}

	//skip the record length
	_, err = buffer.ReadByte()
	if err != nil {
		return err
	}

	binary.Read(buffer, binary.LittleEndian, &r.sdrCompactSensorFields)

	idLen, err := buffer.ReadByte()
	if err != nil {
		return err
	}

	idLen = parseIdLen(idLen)

	id := make([]byte, int(idLen))
	n, err := buffer.Read(id)
	if err != nil || n != int(idLen) {
		return ErrIdStringLenNotMatch
	}

	r.Deviceid = string(id)

	return nil
}

// section 43.9
type sdrMCDeviceLocFields struct { //size 26
	DeviceSlaveAddr uint8
	ChannNum        uint8
	PowerStaNotif   uint8
	DeviceCapab     uint8
	Reserved        [3]byte
	EntityID        uint8
	EntityInstan    uint8
	OEM             SDRSensorType
	DeviceIDCode    SDRSensorReadingType
	DeviceIDStr     uint16
}
type SDRMCDeviceLoc struct {
	SDRRecordHeader
	sdrMCDeviceLocFields
	Deviceid string
}
