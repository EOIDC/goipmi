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
var discreteSensorStatDesc map[uint8]map[uint8]string = map[uint8]map[uint8]string{
	0x02: {0x00: "Transition to Idle", 0x01: "Transition to Active", 0x02: "Transition to Busy"},
	0x03: {0x00: "State Deasserted", 0x01: "State Asserted"},
	0x04: {0x00: "Predictive Failure deasserted", 0x01: "Predictive Failure asserted"},
	0x05: {0x00: "Limit Not Exceeded", 0x01: "Limit Exceeded"},
	0x06: {0x00: "Performance Met", 0x01: "Performance Lags"},
	0x07: {0x00: "Transition to OK", 0x01: "Transition to Non-Critical from OK", 0x02: "Transition to Critical from less servere",
		0x03: "Transition to Non-recoverable from less servere", 0x04: "Transition to Non-Critical from more servere",
		0x05: "Transition to Critical from Non-recoverable", 0x06: "Transition to Non-recoverable", 0x07: "Monitor", 0x08: "Information"},
	0x08: {0x00: "Device Removed/Device Absent", 0x01: "Device Inserted/Device Present"},
	0x09: {0x00: "Device Disabled", 0x01: "Device Enabled"},
	0x0a: {0x00: "transition to Running", 0x01: "transition to In Test", 0x02: "transition to Power Off",
		0x03: "transition to On Line", 0x04: "transition to Off Line", 0x05: "transition to Off Duty",
		0x06: "transition to Degraded", 0x07: "transition to Power Saved", 0x08: "Install Error"},
	0x0b: {0x00: "Fully Redundant",
		0x01: "Redundancy Lost",
		0x02: "Redundancy Degraded",
		0x03: "Non-redundant:Sufficient Resources from Redundant",
		0x04: "Non-Redundant: Sufficient from Insufficient",
		0x05: "Non-Redundant: Insufficient Resources",
		0x06: "Redundancy Degraded from  Fully Redundant",
		0x07: "redundancy Degraded from Non-redundant"},
	0x0c: {0x00: "D0 Power State", 0x01: "D1 Power State", 0x02: "D2 Power State", 0x03: "D3 Power State"},
}
var sensorTypeCodeEvent map[uint8]map[uint8]string = map[uint8]map[uint8]string{
	0x05: {0x00: "General Chassis intrusion", 0x01: "Drive Bay intrusion", 0x02: "I/O Card area intrusion", 0x03: "Processor area intrusion,",
		0x04: "System unplugged from LAN", 0x05: "Unauthorized dock", 0x06: "FAN area intrusion"},
	0x06: {0x00: "Front Panel Lockout Violation attempt", 0x01: "Pre-boot Password Violation- user password",
		0x02: "Pre-boot Password Violation attempt-setup password", 0x03: "Pre-boot Password Violation- network boot password",
		0x04: "Pre-boot Password Violation", 0x05: "Out-of-band Access Password Violation"},
	0x07: {0x00: "IERR", 0x01: "Thermal Trip", 0x02: "FRB1/BIST failure", 0x03: "FRB2/Hang in POST failure",
		0x04: "FRB3/Processor startup/init failfirmure", 0x05: "Configuration Error",
		0x06: "SM BIOS 'Uncorrectable CPU-complex Error'", 0x07: "Processor Presence", 0x08: "disabled",
		0x09: "Terminator Presence Detected", 0x0a: "Throttled", 0x0b: "Machine Check Exception(Uncorrectable)",
		0xc: "Correctable Machine Check Error"},
	0x08: {0x00: "Presence detected", 0x01: "Failure detected", 0x02: "Predictive Failure", 0x03: "Power Supply AC lost",
		0x04: "AC lost or out-of-range", 0x05: "AC out-of-range, but present",
		0x06: "Config Error: Vendor Mismatch",
		0x07: "Power Supply Inative"},
	0x09: {0x00: "Power Off/Down", 0x01: "Power Cycle", 0x02: "240V Power Down", 0x03: "Interlock Power Down",
		0x04: "AC lost", 0x05: "Soft Power Control Failure", 0x06: "Failure detected", 0x07: "Predictive Failure"},
	0x0c: {0x00: "Correctable ECC", 0x01: "UnCorrectable ECC",
		0x02: "Parity", 0x03: "Memory Scrub Failed", 0x04: "Memory Device Disabled", 0x05: "Correctable ECC logging limit reached",
		0x06: "Presence detected.",
		0x07: "Configuration error.", 0x08: "Spare.", 0x09: "Throttled", 0x0a: "Critical Overtemperature"},
	0x0d: {0x00: "Drive Presence", 0x01: "Drive Fault", 0x02: "Predictive Fault", 0x03: "Hot Spare", 0x04: "Parity Check in progress",
		0x05: "In Critical Array", 0x06: "In Failed Array", 0x07: "Rebuild/Remap in progress", 0x08: "Rebuild Aborted"},
	0x0f: {0x00: "System Firmware Error", 0x01: "System Firmware Hang ", 0x02: "System Firmware Progress"},
	0x10: {0x00: "Correctable memory error logging disabled", 0x01: "Event logging disabled", 0x02: "Log area reset/cleared",
		0x03: "All event logging disabled", 0x04: "Log full", 0x05: "Log almost full"},
	0x11: {0x00: "BIOS Reset", 0x01: "OS Reset", 0x02: "OS Shut Down", 0x03: "OS Power Down", 0x04: "OS Power Cycle",
		0x05: "OS NMI/Diag Interrupt", 0x06: "OS Expired", 0x07: "OS pre-timeout Interrupt"},
	0x12: {0x00: "System Reconfigured", 0x01: "OEM System boot event", 0x02: "Undetermined system hardware failure",
		0x03: "Entry added to auxiliary log", 0x04: "PEF Action", 0x05: "Timestamp Clock Sync"},
	0x13: {0x00: "NMI/Diag Interrupt", 0x01: "Bus Timeout", 0x02: "I/O Channel check NMI",
		0x03: "Software NMI", 0x04: "PCI PERR", 0x05: "PCI SERR",
		0x06: "EISA failsafe timeout", 0x07: "Bus Correctable error", 0x08: "Bus Uncorrectable error",
		0x09: "Fatal NMI", 0x0a: "Bus Fatal Error", 0x0b: "Bus Degraded"},
	0x14: {0x00: "Power Button pressed", 0x01: "Sleep Button pressed", 0x02: "Reset Button pressed",
		0x03: "FRU Latch", 0x04: "FRU Service"},
	0x19: {0x00: "Soft Power Control Failure", 0x01: "Thermal Trip"},
	0x1b: {0x1b: "Connected", 0x01: "Config Error"},
	0x1d: {0x00: "Initiated by power up", 0x01: "Initiated by hard reset", 0x02: "Initiated by warm reset",
		0x03: "User requested PXE boot", 0x04: "Automatic boot to diagnostic", 0x05: "OS initiated hard reset",
		0x06: "OS initiated warm reset", 0x07: "System Restart"},
	0x1e: {0x00: "No bootable media", 0x01: "Non-bootable disk in drive", 0x02: "PXE server not found",
		0x03: "Invalid boot sector", 0x04: "Timeout waiting for selection"},
	0x1f: {0x00: "A: boot completed", 0x01: "C: boot completed", 0x02: "PXE boot completed",
		0x03: "Diagnostic boot completed", 0x04: "CD-ROM boot completed", 0x05: "ROM boot completed",
		0x06: "boot completed - device not specified", 0x07: "Installation started", 0x08: "Installation completed",
		0x09: "Installation aborted", 0x0a: "Installation failed"},
	0x20: {0x00: "Error during system startup", 0x01: "Run-time critical stop", 0x02: "OS graceful stop",
		0x03: "OS graceful shutdown", 0x04: "PEF initiated soft shutdown", 0x05: "Agent not responding"},
	0x21: {0x00: "Fault Status", 0x01: "Identify Status", 0x02: "Device Installed",
		0x03: "Ready for Device Installation", 0x04: "Ready for Device Removal", 0x05: "Slot Power is Off",
		0x06: "Device Removal Request", 0x07: "Interlock", 0x08: "Slot is Disabled", 0x09: "Spare Device"},
	0x22: {0x00: "S0/G0: working", 0x01: "S1: sleeping with system hw & processor context maintained",
		0x02: "'S2: sleeping',processor context lost", 0x03: "'S3: sleeping,processor & hw context lost',memory retained",
		0x04: "S4: non-volatile sleep/suspend-to-disk", 0x05: "S5/G2: soft-off", 0x06: "S4/S5: soft-off",
		0x07: "G3: mechanical off", 0x08: "Sleeping in S1/S2/S3 state", 0x09: "G1: sleeping",
		0x0a: "S5: entered by override", 0x0b: "Legacy ON state", 0x0c: "Legacy OFF state", 0x0e: "Unknown"},
	0x23: {0x00: "Timer expired", 0x01: "Hard reset", 0x02: "Power down", 0x03: "Power cycle",
		0x04: "reserved", 0x05: "reserved", 0x06: "reserved", 0x07: "reserved"},
	0x24: {0x00: "Platform generated page", 0x01: "Platform generated LAN alert", 0x02: "Platform Event Trap generated",
		0x03: "Platform generated SNMP trap,OEM format"},
	0x25: {0x00: "Present", 0x01: "Absent", 0x02: "Disabled"},
	0x27: {0x00: "Heartbeat Lost", 0x01: "Heartbeat"},
	0x28: {0x00: "Sensor access degraded or unavailable", 0x01: "Controller access degraded or unavailable",
		0x02: "Management controller off-line", 0x03: "Management controller unavailable",
		0x04: "Sensor failure", 0x05: "FRU failure"},
	0x29: {0x00: "Low", 0x01: "Failed", 0x02: "Presence Detected"},
	0x2a: {0x00: "Session Activated", 0x01: "Session Deactivated", 0x02: "Invalid Username or Password",
		0x03: "Invalid password disable."},
	0x2b: {0x00: "Hardware change detected", 0x01: "Hardware incompatibility detected", 0x03: "Firmware or software incompatibility detected",
		0x04: "Invalid or unsupported hardware version", 0x05: "Invalid or unsupported firmware or software version",
		0x06: "Hardware change success", 0x07: "Firmware or software change success"},
	0x2c: {0x00: "Not Installed", 0x01: "Inactive", 0x02: "Activation Requested", 0x03: "Activation in Progress",
		0x04: "Active", 0x05: "Deactivation Requested", 0x06: "Deactivation in Progress", 0x07: "Communication lost"},
}
var sysFirmwareEvent map[uint8]string = map[uint8]string{
	0x00: "Unspecified", 0x01: "No system memory is physically installed in the system.", 0x02: "No usable system memory, all installed memory " +
		"hasexperienced an unrecoverable failure.", 0x03: "Unrecoverable hard-disk/ATAPI/IDE device failure.", 0x04: "Unrecoverable system-board failure",
	0x05: "Unrecoverable diskette subsystem failure.", 0x06: "Unrecoverable hard-disk controller failure.", 0x07: "Unrecoverable PS/2 or USB keyboard failure.",
	0x08: "Removable boot media not found", 0x09: "Unrecoverable video controller failure", 0x0a: "No video device detected", 0x0b: "Firmware (BIOS) ROM " +
		"corruption detected", 0x0c: "CPU voltage mismatch (processors that share same supply have mismatched voltage requirements)",
	0x0d: "CPU speed matching failure", 0x0e: "to FFh reserved"}

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
// 36.3
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
	return len & 0x1f
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
	Reserved             [3]byte
	OEM                  uint8
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
