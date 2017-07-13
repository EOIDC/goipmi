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
	//"fmt"
	"math"
	"math/rand"
	//"reflect"
	"time"
)

var defaultRepo map[uint16]*repo

const sdrOpSupport = SDR_OP_SUP_RESERVE_REPO

// combine the SDRRecord and a float64 value
type sDRRecordAndValue struct {
	SDRRecord
	value          float64
	avail          bool
	sensorStatDesc string
	sensorEvent    [16]string
}

type repo struct {
	sdrRepo []*sDRRecordAndValue
}

func NewRepo() *repo {
	rep := &repo{}
	rep.sdrRepo = make([]*sDRRecordAndValue, 0)
	return rep
}
func (rep *repo) initRepoData() {

	//this record is used to unit test
	r1, _ := NewSDRFullSensor(1, "Ambient Temp")
	r1.Unit = 0x00
	r1.BaseUnit = 0x12
	r1.EntityId = 0x37
	r1.EntityIns = 0x01
	r1.SensorType = SDR_SENSOR_TYPECODES_PROCESSOR
	r1.BaseUnit = 0x01

	r1.SensorNumber = 0x04
	r1.ReadingType = SENSOR_READTYPE_THREADHOLD
	r1.SetMBExp(63, 0, 0, 0)
	rep.addRecord(&sDRRecordAndValue{
		SDRRecord: r1,
		value:     2583.0,
		avail:     true,
		sensorStatDesc:"test",
		sensorEvent:[16]string{""},
	})

	r2, _ := NewSDRFullSensor(2, "CPU1 DTS")
	r2.Unit = 0x80
	r2.SensorNumber = 0x05
	r2.ReadingType = SENSOR_READTYPE_THREADHOLD
	r2.BaseUnit = 0x01
	r2.SetMBExp(1, 0, 0, 0)
	rep.addRecord(&sDRRecordAndValue{
		SDRRecord: r2,
		value:     -49.0,
		avail:     true,
		sensorStatDesc:"test",
		sensorEvent:[16]string{""},
	})
	//todo 测试更多类型的sensorRecord
	//	r4, _ := NewSDRFullSensor(2, "CPU1 DTS")
	//	r4.Unit = 0x80
	//	r4.SensorNumber = 0x08
	//	r4.ReadingType = SENSOR_READTYPE_GENERIC_L
	//	r4.BaseUnit = 0x01
	//	r4.SetMBExp(1, 0, 0, 0)
	//	rep.addRecord(&sDRRecordAndValue{
	//		SDRRecord: r4,
	//		value:     -49.0,
	//		avail:      true,
	//	})

	//	r3, _ := NewSDRCompactSensor(2, "CPU1 DTS")
	//	rep.addRecord(&sDRRecordAndValue{
	//		SDRRecord: r3,
	//		value:     -49.0,
	//		avail:      false,
	//	})
}

func (rep *repo) addRecord(rec *sDRRecordAndValue) {
	rep.sdrRepo = append(rep.sdrRepo, rec)
}

// return the marshaled bytes and next record id
func (rep *repo) getRecordById(id uint16) ([]byte, uint16) {

	if len(rep.sdrRepo) < 1 {
		return nil, 0xffff
	}

	//if not match, try to find the nearest one
	var distance = uint16(0xffff)
	var index = 0
	var record *sDRRecordAndValue
	var next uint16
	var data []byte
	for i, rec := range rep.sdrRepo {
		_distance := uint16(math.Abs(float64(rec.RecordId()) - float64(id)))
		if _distance < distance {
			distance = _distance
			index = i
		}
	}

	if index >= len(rep.sdrRepo)-1 {
		next = 0xffff
	} else {
		next = rep.sdrRepo[index+1].RecordId()
	}

	record = rep.sdrRepo[index]

	switch record.RecordType() {
	case SDR_RECORD_TYPE_FULL_SENSOR:
		r, _ := record.SDRRecord.(*SDRFullSensor)
		data, _ = r.MarshalBinary()
		break
	case SDR_RECORD_TYPE_MC_DEVICE_LOCATOR:
		r, _ := record.SDRRecord.(*SDRMcDeviceLocator)
		data, _ = r.MarshalBinary()
		break
	default:
		panic("Unsupport record type")
	}

	return data, next
}

func (s *Simulator) repositoryInfo(*Message) Response {
	return &SDRRepositoryInfoResponse{
		CompletionCode:              CommandCompleted,
		SDRVersion:                  0x51,
		RecordCount:                 10,
		FreeSpaceInBytes:            0,
		TimestampMostRecentAddition: 0,
		TimestampMostRecentErase:    0,
		OperationSupprot:            sdrOpSupport,
	}
}

func (s *Simulator) reserveRepository(*Message) Response {

	var rId uint16
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	rId = uint16(r.Intn(65536))

	if defaultRepo == nil {
		defaultRepo = make(map[uint16]*repo)
	}
	defaultRepo[rId] = NewRepo()
	defaultRepo[rId].initRepoData()

	return &ReserveRepositoryResponse{
		CompletionCode: CommandCompleted,
		ReservationId:  rId,
	}
}

func (s *Simulator) getSDR(m *Message) Response {
	request := &GetSDRCommandRequest{}
	var rep *repo
	var ok bool
	if err := m.Request(request); err != nil {
		return err
	}
	rId := request.ReservationID
	if rep, ok = defaultRepo[rId]; !ok {
		//TODO return err
		panic("rId not found")
	}

	data, nid := rep.getRecordById(request.RecordID)
	response := &GetSDRCommandResponse{}
	response.CompletionCode = CommandCompleted
	response.NextRecordID = nid
	response.ReadData = data[request.OffsetIntoRecord : request.OffsetIntoRecord+request.ByteToRead]
	return response
}
func (s *Simulator) getSensorReading(m *Message) Response {
	request := &GetSensorReadingRequest{}
	if err := m.Request(request); err != nil {
		return err
	}
	sensorNum := request.SensorNumber
	var rep *sDRRecordAndValue = nil
	for _, value := range defaultRepo {
		for _, sdrRepo2 := range value.sdrRepo {
			sdrFullSensor := (sdrRepo2.SDRRecord).(*SDRFullSensor)
			if sdrFullSensor.SensorNumber == sensorNum {
				rep = sdrRepo2
				break
			}
		}
	}
	if rep == nil {
		return nil
	} else {
		response := &GetSensorReadingResponse{}
		value := rep.value
		if sdrFullSensor2, ok := (rep.SDRRecord).(*SDRFullSensor); ok {
			sensorReading2 := sdrFullSensor2.CalValue(value)
			response.SensorReading = sensorReading2
		} else {
			response.SensorReading = uint8(value)
		}
		response.CompletionCode = CommandCompleted
		fmt.Println("rep.avail==", rep.avail)
		if rep.avail == true {
			//response.ReadingAvail & 0x20 != 0    (11 10 01 00) 000000
			response.ReadingAvail = 0x40
		} else {
			//response.ReadingAvail & 0x20 = 0 (11 10 01 00) 100000
			response.ReadingAvail = 0x20
		}
		return response
	}
}
