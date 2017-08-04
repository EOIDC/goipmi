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

// section 33.9
const (
	SDR_OP_SUP_ALLOC_INFO   = (1 << 0)
	SDR_OP_SUP_RESERVE_REPO = (1 << 1)
	SDR_OP_SUP_PARTIAL_ADD  = (1 << 2)
	SDR_OP_SUP_DELETE       = (1 << 3)
	SDR_OP_SUP_NON_MODAL_UP = (1 << 5)
	SDR_OP_SUP_MODAL_UP     = (1 << 6)
	SDR_OP_SUP_OVERFLOW     = (1 << 7)
)

type SDRRecordType uint8

const (
	SDR_RECORD_TYPE_FULL_SENSOR            = 0x01
	SDR_RECORD_TYPE_COMPACT_SENSOR         = 0x02
	SDR_RECORD_TYPE_EVENTONLY_SENSOR       = 0x03
	SDR_RECORD_TYPE_ENTITY_ASSOC           = 0x08
	SDR_RECORD_TYPE_DEVICE_ENTITY_ASSOC    = 0x09
	SDR_RECORD_TYPE_GENERIC_DEVICE_LOCATOR = 0x10
	SDR_RECORD_TYPE_FRU_DEVICE_LOCATOR     = 0x11
	SDR_RECORD_TYPE_MC_DEVICE_LOCATOR      = 0x12
	SDR_RECORD_TYPE_MC_CONFIRMATION        = 0x13
	SDR_RECORD_TYPE_BMC_MSG_CHANNEL_INFO   = 0x14
	SDR_RECORD_TYPE_OEM                    = 0xc0
)

type SDRSensorType uint8

const (
	RESERVED                                  = 0x00
	SDR_SENSOR_TYPECODES_TEMPERATURE          = 0x01
	SDR_SENSOR_TYPECODES_VOLTAGE              = 0x02
	SDR_SENSOR_TYPECODES_CURRENT              = 0x03
	SDR_SENSOR_TYPECODES_FAN                  = 0x04
	SDR_SENSOR_TYPECODES_PHYSICALSECURITY     = 0x05
	SDR_SENSOR_TYPECODES_PALTFORMSECURITY     = 0x06
	SDR_SENSOR_TYPECODES_PROCESSOR            = 0x07
	SDR_SENSOR_TYPECODES_POWSERSUPPLY         = 0x08
	SDR_SENSOR_TYPECODES_POWERUNIT            = 0x09
	SDR_SENSOR_TYPECODES_COLLINGDEVICE        = 0x0a
	SDR_SENSOR_TYPECODES_OTHERUNITSBASED      = 0x0b
	SDR_SENSOR_TYPECODES_MEMORY               = 0x0c
	SDR_SENSOR_TYPECODES_DRIVESLOT            = 0x0d
	SDR_SENSOR_TYPECODES_POSTMEMORYRESIZE     = 0x0e
	SDR_SENSOR_TYPECODES_SYSTEMFIRMWARE       = 0x0f
	SDR_SENSOR_TYPECODES_EVENTLOGGINGDISABLED = 0x10
	SDR_SENSOR_TYPECODES_WATCHDOG1            = 0x11
	SDR_SENSOR_TYPECODES_SYSTEMEVENT          = 0x12
	SDR_SENSOR_TYPECODES_CRITICALINTERRUPT    = 0x13
	SDR_SENSOR_TYPECODES_BUTTONSWITCH         = 0x14
	SDR_SENSOR_TYPECODES_MODULEBOARD          = 0x15
	SDR_SENSOR_TYPECODES_MICROCTRLCOPROCESS   = 0x16
	SDR_SENSOR_TYPECODES_ADDINCARD            = 0x17
	SDR_SENSOR_TYPECODES_CHASSIS              = 0x18
	SDR_SENSOR_TYPECODES_CHIPSET              = 0x19
	SDR_SENSOR_TYPECODES_OTHER_FRU            = 0x1a
	SDR_SENSOR_TYPECODES_CABLEINTERCONN       = 0x1b
	SDR_SENSOR_TYPECODES_TERMINATOR           = 0x1c
	SDR_SENSOR_TYPECODES_SYSBOOT_RESTAINITI   = 0x1d
	SDR_SENSOR_TYPECODES_BOOTERROR            = 0x1e
	SDR_SENSOR_TYPECODES_BASEOSBOOT_INITISTAT = 0x1f
	SDR_SENSOR_TYPECODES_OSSTOP_SHUTDOWN      = 0x20
	SDR_SENSOR_TYPECODES_STOPCONNECTOR        = 0x21
	SDR_SENSOR_TYPECODES_SYSTEMACPIPOWERSTAT  = 0x22
	SDR_SENSOR_TYPECODES_WATCHDOG2            = 0x23
	SDR_SENSOR_TYPECODES_PLATORMALERT         = 0x24
	SDR_SENSOR_TYPECODES_ENTITYPRESENCE       = 0x25
	SDR_SENSOR_TYPECODES_MONITORASIC_IC       = 0x26
	SDR_SENSOR_TYPECODES_LAN                  = 0x27
	SDR_SENSOR_TYPECODES_MANAGSUBSYSHEALTH    = 0x28
	SDR_SENSOR_TYPECODES_BATTERY              = 0x29
	SDR_SENSOR_TYPECODES_SESSIONAUDIT         = 0x2a
	SDR_SENSOR_TYPECODES_VERSIONCHANGE        = 0x2b
	SDR_SENSOR_TYPECODES_FRUSTATE             = 0x2c
)

type SDRSensorReadingType uint8

const (
	SENSOR_READTYPE_UNSPECIFIED  = 0x00
	SENSOR_READTYPE_THREADHOLD   = 0x01
	SENSOR_READTYPE_GENERIC_L    = 0x02
	SENSOR_READTYPE_GENERIC_H    = 0x0C
	SENSOR_READTYPE_SENSORSPECIF = 0x6F
	SENSOR_READTYPE_OEM_L        = 0x70
	SENSOR_READTYPE_OEM_H        = 0x7F
)

type SDRSensorReadingData2 uint8

const (
	SDR_SENSOR_STAT_LO_NC = 1 << 0
	SDR_SENSOR_STAT_LO_CR = 1 << 1
	SDR_SENSOR_STAT_LO_NR = 1 << 2
	SDR_SENSOR_STAT_HI_NC = 1 << 3
	SDR_SENSOR_STAT_HI_CR = 1 << 4
	SDR_SENSOR_STAT_HI_NR = 1 << 5
)
