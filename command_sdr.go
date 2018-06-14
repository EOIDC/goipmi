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
)

const (
	CommandGetSDRRepositoryInfo = Command(0x20)
	CommandGetReserveSDRRepo    = Command(0x22)
	CommandGetSDR               = Command(0x23)
	CommandGetSensorReading     = Command(0x2d)
)

type ReserveSDRRepositoryRequest struct{}

// section 33.11
type ReserveRepositoryResponse struct {
	CompletionCode
	ReservationId uint16
}
type GetSDRCommandRequest struct {
	ReservationID    uint16
	RecordID         uint16
	OffsetIntoRecord uint8
	ByteToRead       uint8 //FFH means read entire record
}
type GetSDRCommandResponse struct {
	CompletionCode
	NextRecordID uint16
	ReadData     []byte
}
type GetSensorReadingRequest struct {
	SensorNumber uint8
}

// section 35.14
type GetSensorReadingResponse struct {
	CompletionCode
	SensorReading uint8
	ReadingAvail  uint8
	Data1         uint8
	Data2         uint8
}

func (r *GetSDRCommandResponse) MarshalBinary() (data []byte, err error) {
	buffer := new(bytes.Buffer)
	binary.Write(buffer, binary.LittleEndian, r.CompletionCode)
	binary.Write(buffer, binary.LittleEndian, r.NextRecordID)
	buf := make([]byte, 0)
	buf = append(buf, buffer.Bytes()...)
	buf = append(buf, r.ReadData...)
	return buf, nil
}

func (r *GetSDRCommandResponse) UnmarshalBinary(data []byte) error {
	buffer := bytes.NewBuffer(data)

	var cc CompletionCode
	var nrid uint16
	var err error
	err = binary.Read(buffer, binary.LittleEndian, &cc)
	if err != nil {
		return err
	}
	err = binary.Read(buffer, binary.LittleEndian, &nrid)
	if err != nil {
		return err
	}
	r.CompletionCode = cc
	r.NextRecordID = nrid
	r.ReadData = data[3:]
	return nil
}

// 35.14 Data2 is Optional, so we have to implement UnmarshalBinary manually
func (r *GetSensorReadingResponse) UnmarshalBinary(data []byte) error {
	buffer := bytes.NewBuffer(data)
	var err error
	err = binary.Read(buffer, binary.LittleEndian, &r.CompletionCode)
	if err != nil {
		return err
	}

	err = binary.Read(buffer, binary.LittleEndian, &r.SensorReading)
	if err != nil {
		return err
	}

	err = binary.Read(buffer, binary.LittleEndian, &r.ReadingAvail)
	if err != nil {
		return err
	}

	err = binary.Read(buffer, binary.LittleEndian, &r.Data1)
	if err != nil {
		return err
	}

	r.Data2 = uint8(0)
	binary.Read(buffer, binary.LittleEndian, &r.Data2)
	return nil
}
