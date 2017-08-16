package ipmi

import (
	"bytes"
	"errors"
	"fmt"
	"math"
)

type SdrSensorInfo struct {
	SensorType  string
	ReadingType SDRSensorReadingType
	BaseUnit    string
	Value       float64
	DeviceId    string
	StatusDesc  string
	SensorEvent []string
	Avail       bool
	Data1       uint8
	Data2       uint8
}

// RepositoryInfo get the Repository Info of the SDR
func (c *Client) RepositoryInfo() (*SDRRepositoryInfoResponse, error) {
	req := &Request{
		NetworkFunctionStorge,
		CommandGetSDRRepositoryInfo,
		&SDRRepositoryInfoRequest{},
	}
	res := &SDRRepositoryInfoResponse{}
	return res, c.Send(req, res)
}
func (c *Client) GetReserveSDRRepoForReserveId() (*ReserveRepositoryResponse, error) {
	req := &Request{
		NetworkFunctionStorge,
		CommandGetReserveSDRRepo,
		&ReserveSDRRepositoryRequest{},
	}
	res := &ReserveRepositoryResponse{}
	return res, c.send(req, res)
}
func GetSensorStatDesc(readingType SDRSensorReadingType, sensorType SDRSensorType, state2 uint8, state3 uint8) (string, []string) {
	sensorStatDescStr := ""
	sensorEven := []string{}
	//sensorEven := make([]string,0,16)
	if readingType == SENSOR_READTYPE_THREADHOLD {
		if (state2 & SDR_SENSOR_STAT_LO_NR) != 0 {
			sensorStatDescStr = "Lower Non-Recoverable"
		} else if (state2 & SDR_SENSOR_STAT_HI_NR) != 0 {
			sensorStatDescStr = "Upper Non-Recoverable"
		} else if (state2 & SDR_SENSOR_STAT_LO_CR) != 0 {
			sensorStatDescStr = "Lower Critical"
		} else if (state2 & SDR_SENSOR_STAT_HI_CR) != 0 {
			sensorStatDescStr = "Upper Critical"
		} else if (state2 & SDR_SENSOR_STAT_LO_NC) != 0 {
			sensorStatDescStr = "Lower Non-Critical"
		} else if (state2 & SDR_SENSOR_STAT_HI_NC) != 0 {
			sensorStatDescStr = "Upper Non-Critical"
		} else {
			sensorStatDescStr = "ok"
		}
	} else if readingType >= SENSOR_READTYPE_GENERIC_L && readingType <= SENSOR_READTYPE_GENERIC_H {
		var i uint8 = 0
		for ; i < 8; i++ {
			if (state2 & (1 << i)) != 0 {
				sensorEven = append(sensorEven, (discreteSensorStatDesc[uint8(readingType)])[i])
			}
		}
		for i = 0; i < 8; i++ {
			if (state3 & (1 << i)) != 0 {
				sensorEven = append(sensorEven, (discreteSensorStatDesc[uint8(readingType)])[i+8])
			}
		}
	} else if readingType == SENSOR_READTYPE_SENSORSPECIF {
		var i uint8 = 0
		for ; i < 8; i++ {
			if (state2 & (1 << i)) != 0 {
				sensorEven = append(sensorEven, (sensorTypeCodeEvent[uint8(sensorType)])[i])
			}
		}
		i = 0
		for ; i < 8; i++ {
			if (state3 & (1 << i)) != 0 {
				sensorEven = append(sensorEven, (sensorTypeCodeEvent[uint8(sensorType)])[i+8])
			}
		}
	}
	return sensorStatDescStr, sensorEven
}

func (c *Client) GetSensorList(reservationID uint16) ([]SdrSensorInfo, error) {
	var recordId uint16 = 0
	var sdrSensorInfolist = make([]SdrSensorInfo, 0, 30)
	for recordId < 0xffff {
		sdrRecordAndValue, nId, err := c.GetSDR(reservationID, recordId)
		if err != nil {
			recordId = nId
			continue
		}
		if fullSensor, ok1 := sdrRecordAndValue.SDRRecord.(*SDRFullSensor); ok1 {
			if fullSensor.BaseUnit >= 0 && fullSensor.BaseUnit < uint8(len(sdrRecordValueBasicUnit)) &&
				fullSensor.SensorType >= 0 && uint8(fullSensor.SensorType) < uint8(len(sdrRecordValueSensorType)) {

				sdrSensorInfolist = append(sdrSensorInfolist, SdrSensorInfo{
					sdrRecordValueSensorType[fullSensor.SensorType],
					fullSensor.ReadingType,
					sdrRecordValueBasicUnit[fullSensor.BaseUnit],
					sdrRecordAndValue.value,
					fullSensor.Deviceid,
					sdrRecordAndValue.sensorStatDesc,
					sdrRecordAndValue.sensorEvent,
					sdrRecordAndValue.avail,
					sdrRecordAndValue.data1,
					sdrRecordAndValue.data2,
				})
			}
		} else if compactSensor, ok2 := sdrRecordAndValue.SDRRecord.(*SDRCompactSensor); ok2 {
			if compactSensor.BaseUnit >= 0 && compactSensor.BaseUnit < uint8(len(sdrRecordValueBasicUnit)) &&
				compactSensor.SensorType >= 0 && uint8(compactSensor.SensorType) < uint8(len(sdrRecordValueSensorType)) {
				sdrSensorInfolist = append(sdrSensorInfolist, SdrSensorInfo{
					sdrRecordValueSensorType[compactSensor.SensorType],
					compactSensor.ReadingType,
					sdrRecordValueBasicUnit[compactSensor.BaseUnit],
					sdrRecordAndValue.value,
					compactSensor.Deviceid,
					sdrRecordAndValue.sensorStatDesc,
					sdrRecordAndValue.sensorEvent,
					sdrRecordAndValue.avail,
					sdrRecordAndValue.data1,
					sdrRecordAndValue.data2,
				})
			}
		}
		recordId = nId
	}
	return sdrSensorInfolist, nil
}

//Get SDR Command  33.12
func (c *Client) GetSDR(reservationID uint16, recordID uint16) (sdr *sDRRecordAndValue, next uint16, err error) {
	req_step1 := &Request{
		NetworkFunctionStorge,
		CommandGetSDR,
		&GetSDRCommandRequest{
			ReservationID:    reservationID,
			RecordID:         recordID,
			OffsetIntoRecord: 0,
			ByteToRead:       5,
		},
	}
	recordKeyBody_Data := new(bytes.Buffer)
	res_step1 := &GetSDRCommandResponse{}
	c.Send(req_step1, res_step1)
	readData_step1 := res_step1.ReadData
	recordType := readData_step1[3]
	lenToRead_step2 := readData_step1[4]
	recordKeyBody_Data.Write(readData_step1)
	req_step2 := &Request{
		NetworkFunctionStorge,
		CommandGetSDR,
		&GetSDRCommandRequest{
			ReservationID:    reservationID,
			RecordID:         recordID,
			OffsetIntoRecord: 5,
			ByteToRead:       uint8(lenToRead_step2),
		},
	}
	res_step2 := &GetSDRCommandResponse{}
	c.Send(req_step2, res_step2)
	recordKeyBody_Data.Write(res_step2.ReadData)
	sdrRecordAndValue, err := c.CalSdrRecordValue(recordType, recordKeyBody_Data)
	return sdrRecordAndValue, res_step2.NextRecordID, err
}
func (c *Client) CalSdrRecordValue(recordType uint8, recordKeyBody_Data *bytes.Buffer) (*sDRRecordAndValue, error) {
	var sdrRecordAndValue = &sDRRecordAndValue{}
	if recordType == SDR_RECORD_TYPE_FULL_SENSOR {
		//Unmarshalbinary and assert
		fullSensor, _ := NewSDRFullSensor(0, "")
		fullSensor.UnmarshalBinary(recordKeyBody_Data.Bytes())
		sdrRecordAndValue.SDRRecord = fullSensor
		sensorReadingRes, err := c.getSensorReading(fullSensor.SensorNumber)
		if err != nil || (sensorReadingRes.ReadingAvail&0x20) > 0 {
			sdrRecordAndValue.avail = false
			sdrRecordAndValue.value = 0.00
			sdrRecordAndValue.sensorStatDesc = "Not Available"
			sdrRecordAndValue.sensorEvent = []string{""}
		} else {
			res, avai := calFullSensorValue(fullSensor, sensorReadingRes.SensorReading)
			sdrRecordAndValue.sensorStatDesc, sdrRecordAndValue.sensorEvent = GetSensorStatDesc(fullSensor.ReadingType, fullSensor.SensorType, sensorReadingRes.Data1, sensorReadingRes.Data2)
			sdrRecordAndValue.avail = avai
			sdrRecordAndValue.value = res
		}
		return sdrRecordAndValue, err
	} else if recordType == SDR_RECORD_TYPE_COMPACT_SENSOR {
		//Unmarshalbinary and assert
		compactSensor, _ := NewSDRCompactSensor(0, "")
		compactSensor.UnmarshalBinary(recordKeyBody_Data.Bytes())
		sdrRecordAndValue.SDRRecord = compactSensor
		sensorReadingRes, err := c.getSensorReading(compactSensor.SensorNumber)
		if err != nil || (sensorReadingRes.ReadingAvail&0x20) > 0 {
			sdrRecordAndValue.avail = false
			sdrRecordAndValue.value = 0.00
			sdrRecordAndValue.sensorStatDesc = ""
			sdrRecordAndValue.sensorEvent = []string{""}
			sdrRecordAndValue.data1 = sensorReadingRes.Data1
			sdrRecordAndValue.data2 = sensorReadingRes.Data2
		} else {
			res, avai := calCompactSensorValue(compactSensor, sensorReadingRes.SensorReading)
			sdrRecordAndValue.sensorStatDesc, sdrRecordAndValue.sensorEvent = GetSensorStatDesc(compactSensor.ReadingType, compactSensor.SensorType, sensorReadingRes.Data1, sensorReadingRes.Data2)
			sdrRecordAndValue.avail = avai
			sdrRecordAndValue.value = res
			sdrRecordAndValue.data1 = sensorReadingRes.Data1
			sdrRecordAndValue.data2 = sensorReadingRes.Data2
		}
		return sdrRecordAndValue, err
	} else {
		return nil, errors.New(fmt.Sprintf("Unsupport Record Type %d", recordType))
	}
	return nil, nil
}
func calFullSensorValue(sdrRecord SDRRecord, sensorReading uint8) (float64, bool) {
	if fullSensor, err := sdrRecord.(*SDRFullSensor); err {
		var result float64 = 0.0
		var avail bool
		//threshold type
		if fullSensor.ReadingType == SENSOR_READTYPE_THREADHOLD {
			// has analog value
			if fullSensor.Unit&0xc0 != 0xc0 {
				m, b, bexp, rexp := fullSensor.GetMBExp()
				switch (fullSensor.Unit & 0xc0) >> 6 {
				case 0:
					result = (float64(m)*float64(sensorReading) + float64(b)*math.Pow(10, float64(bexp))) * math.Pow(10, float64(rexp))
					break
				case 1:
					if (sensorReading & 0x80) == 0 {
						sensorReading++
					}
				case 2:
					result = (float64(int8(m)*int8(sensorReading)) + float64(b)*math.Pow(10, float64(rexp))) * math.Pow(10, float64(bexp))
					break
				}
				avail = true
			} else {
				avail = false
			}
		}
		return result, avail
	}

	return float64(0), false
}
func calCompactSensorValue(sdrRecord SDRRecord, sensorReading uint8) (float64, bool) {
	var value float64 = 0.0
	var avail bool = false
	if compactSensor, err := sdrRecord.(*SDRCompactSensor); err {
		//threshold type
		if compactSensor.ReadingType == SENSOR_READTYPE_THREADHOLD {
			// has analog value
			if compactSensor.Unit&0xc0 == 0xc0 {
				avail = true
				value = float64(sensorReading)
			} else {
				avail = false
				value = 0.0
			}
		} else if compactSensor.ReadingType == SENSOR_READTYPE_SENSORSPECIF {
			// has analog value
			if compactSensor.Unit&0xc0 == 0xc0 {
				avail = true
				value = float64(sensorReading)
			} else {
				avail = false
				value = 0.0
			}
		} else if compactSensor.ReadingType >= SENSOR_READTYPE_GENERIC_L && compactSensor.ReadingType <= SENSOR_READTYPE_GENERIC_L {
			// has analog value
			if compactSensor.Unit&0xc0 == 0xc0 {
				avail = true
				value = float64(sensorReading)
			} else {
				avail = false
				value = 0.0
			}
		}
	}
	return value, avail
}

//Get Sensor Reading  35.14
func (c *Client) getSensorReading(sensorNum uint8) (*GetSensorReadingResponse, error) {
	req := &Request{
		NetworkFunctionSensorEvent,
		CommandGetSensorReading,
		&GetSensorReadingRequest{
			SensorNumber: sensorNum,
		},
	}
	res := &GetSensorReadingResponse{}
	c.Send(req, res)
	if res == nil {
		return nil, ErrNotFoundTheSensorNum
	}
	//if (res.ReadingAvail & 0x20) == 0 {
	//	readValue := res.SensorReading
	//	return readValue, nil
	//}
	//return uint8(0), ErrSensorReadUnavail
	return res, nil
}
