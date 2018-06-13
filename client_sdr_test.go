package ipmi

//
import (
	"github.com/stretchr/testify/assert"
	"net"
	"testing"
)

func TestRepositoryInfo(t *testing.T) {
	s := NewSimulator(net.UDPAddr{})
	err := s.Run()
	assert.NoError(t, err)

	client, err := NewClient(s.NewConnection())
	assert.NoError(t, err)

	err = client.Open()
	assert.NoError(t, err)

	s.SetHandler(NetworkFunctionStorge, CommandGetSDRRepositoryInfo, func(*Message) Response {
		return &SDRRepositoryInfoResponse{
			CompletionCode:   CommandCompleted,
			SDRVersion:       0x51,
			RecordCount:      55,
			OperationSupprot: SDR_OP_SUP_ALLOC_INFO,
		}
	})

	resp, err := client.RepositoryInfo()
	assert.NoError(t, err)
	assert.Equal(t, CommandCompleted, resp.CompletionCode)
	assert.Equal(t, uint16(55), resp.RecordCount)
	assert.Equal(t, uint8(SDR_OP_SUP_ALLOC_INFO), resp.OperationSupprot)

	err = client.Close()
	assert.NoError(t, err)
	s.Stop()

}
func TestGetReserveSDRRepoForReserveId1(t *testing.T) {
	s := NewSimulator(net.UDPAddr{})
	err := s.Run()
	assert.NoError(t, err)

	client, err := NewClient(s.NewConnection())
	assert.NoError(t, err)

	err = client.Open()
	assert.NoError(t, err)

	s.SetHandler(NetworkFunctionStorge, CommandGetReserveSDRRepo, func(*Message) Response {
		return &ReserveRepositoryResponse{
			CompletionCode: CommandCompleted,
			ReservationId:  25555,
		}
	})

	resp, err := client.GetReserveSDRRepoForReserveId()
	assert.NoError(t, err)
	assert.Equal(t, CommandCompleted, resp.CompletionCode)
	assert.Equal(t, uint16(25555), resp.ReservationId)

	err = client.Close()
	assert.NoError(t, err)
	s.Stop()
}

func TestGetSDR(t *testing.T) {
	s := NewSimulator(net.UDPAddr{})
	resp := s.reserveRepository(nil)
	reserve, _ := resp.(*ReserveRepositoryResponse)

	err := s.Run()
	assert.NoError(t, err)

	client, err := NewClient(s.NewConnection())
	assert.NoError(t, err)

	err = client.Open()
	assert.NoError(t, err)

	r1 := &SDRFullSensor{}
	r1.Recordid = 5
	r1.Rtype = SDR_RECORD_TYPE_FULL_SENSOR
	r1.SDRVersion = 0x51
	r1.Deviceid = "fullsensor"
	r1.Unit = 0x00
	r1.SensorNumber = 0x04
	r1.SensorType = SDR_SENSOR_TYPECODES_PROCESSOR
	r1.BaseUnit = 0x12
	r1.SetMBExp(63, 0, 0, 0)
	data1, _ := r1.MarshalBinary()

	response := &GetSDRCommandResponse{}
	response.CompletionCode = CommandCompleted
	response.NextRecordID = 65535

	s.SetHandler(NetworkFunctionStorge, CommandGetSDR, func(m *Message) Response {
		request := &GetSDRCommandRequest{}
		if err := m.Request(request); err != nil {
			return err
		}
		response.ReadData = data1[request.OffsetIntoRecord : request.OffsetIntoRecord+request.ByteToRead]
		return response
	})

	sdrRecordAndValue1, nextRecordId1, err1 := client.GetSDR(reserve.ReservationId, 0)
	if err1 == nil {
		r11 := sdrRecordAndValue1.SDRRecord.(*SDRFullSensor)
		assert.Equal(t, SDRRecordType(SDR_RECORD_TYPE_FULL_SENSOR), r11.Rtype)
		assert.Equal(t, "fullsensor", r11.DeviceId())
		assert.Equal(t, uint16(65535), nextRecordId1)
	}

	r2 := &SDRCompactSensor{}
	r2.Recordid = 10
	r2.Rtype = SDR_RECORD_TYPE_COMPACT_SENSOR
	r2.SDRVersion = 0x51
	r2.Deviceid = "compactsensor"
	r2.Unit = 0x00
	r2.SensorNumber = 0x04
	r2.BaseUnit = 0x12
	data2, _ := r2.MarshalBinary()

	response.CompletionCode = CommandCompleted
	response.NextRecordID = 65535

	s.SetHandler(NetworkFunctionStorge, CommandGetSDR, func(m *Message) Response {
		request := &GetSDRCommandRequest{}
		if err := m.Request(request); err != nil {
			return err
		}
		response.ReadData = data2[request.OffsetIntoRecord : request.OffsetIntoRecord+request.ByteToRead]
		return response
	})
	sdrRecordAndValue2, nextRecordId2, err2 := client.GetSDR(reserve.ReservationId, 0)
	if err2 == nil {
		r22 := sdrRecordAndValue2.SDRRecord.(*SDRCompactSensor)
		assert.Equal(t, SDRRecordType(SDR_RECORD_TYPE_COMPACT_SENSOR), r22.Rtype)
		assert.Equal(t, "compactsensor", r22.DeviceId())
		assert.Equal(t, uint16(65535), nextRecordId2)
	}

	err = client.Close()
	assert.NoError(t, err)
	s.Stop()

}

func TestCalFullSensorValue(t *testing.T) {
	fs1 := &SDRFullSensor{}
	fs1.SetMBExp(8, 0, 0, 0)
	fs1.ReadingType = SENSOR_READTYPE_THREADHOLD
	fs1.Unit = 0x0
	res, Avail := calFullSensorValue(fs1, 0x11)
	assert.Equal(t, float64(136.0), res)
	assert.Equal(t, true, Avail)

	fs1.SetMBExp(1, 0, 0, 0)
	fs1.ReadingType = SENSOR_READTYPE_THREADHOLD
	fs1.Unit = 0x80
	res, Avail = calFullSensorValue(fs1, 0xcf)
	assert.Equal(t, float64(-49.0), res)
	assert.Equal(t, true, Avail)

	fs1.SetMBExp(2, 0, 0, -2)
	fs1.ReadingType = SENSOR_READTYPE_THREADHOLD
	fs1.Unit = 0x00
	res, Avail = calFullSensorValue(fs1, 0xa8)
	assert.Equal(t, float64(3.36), res)

	fs1.SetMBExp(2, 0, 0, -2)
	fs1.ReadingType = SENSOR_READTYPE_THREADHOLD
	fs1.Unit = 0x00
	res, Avail = calFullSensorValue(fs1, 0xa8)
	assert.Equal(t, float64(3.36), res)

	assert.Equal(t, true, Avail)
}
func TestCalCompactSensorValue(t *testing.T) {
	cs1 := &SDRCompactSensor{}
	cs1.ReadingType = SENSOR_READTYPE_SENSORSPECIF
	cs1.Unit = 0xc0
	res, avail := calCompactSensorValue(cs1, 0x11)
	assert.Equal(t, float64(17), res)
	assert.Equal(t, false, avail)
}
func TestGetSensorReading(t *testing.T) {

	s := NewSimulator(net.UDPAddr{})

	err := s.Run()
	assert.NoError(t, err)

	client, err := NewClient(s.NewConnection())
	assert.NoError(t, err)

	s.SetHandler(NetworkFunctionSensorEvent, CommandGetSensorReading, func(m *Message) Response {
		return &GetSensorReadingResponse{
			CompletionCode: CommandCompleted,
			SensorReading:  56,
			ReadingAvail:0xc0,
			Data1:0xc0,
			Data2:0x00,
		}
	})

	err = client.Open()
	assert.NoError(t, err)

	if res, err := client.getSensorReading(0x04); err == nil {
		assert.Equal(t, uint8(56), res.SensorReading)
	}
}

//用与实际采集的数据完全相同进行测试
func TestGetSensorList(t *testing.T) {

	s := NewSimulator(net.UDPAddr{})
	resp := s.reserveRepository(nil)
	reserve, _ := resp.(*ReserveRepositoryResponse)

	err := s.Run()
	assert.NoError(t, err)

	client, err := NewClient(s.NewConnection())
	assert.NoError(t, err)

	err = client.Open()
	assert.NoError(t, err)

	r1 := &SDRFullSensor{}
	r1.Recordid = 5
	r1.Rtype = SDR_RECORD_TYPE_FULL_SENSOR
	r1.SDRVersion = 0x51
	r1.Deviceid = "Fan 5"
	r1.Unit = 0x0
	r1.SensorNumber = 0x04
	r1.SensorType = SDR_SENSOR_TYPECODES_FAN
	r1.BaseUnit = 0x12
	r1.SetMBExp(63, 0, 0, 0)
	r1.ReadingType = SENSOR_READTYPE_THREADHOLD
	data1, _ := r1.MarshalBinary()

	response := &GetSDRCommandResponse{}
	response.CompletionCode = CommandCompleted
	response.NextRecordID = 0xffff

	s.SetHandler(NetworkFunctionStorge, CommandGetSDR, func(m *Message) Response {
		request := &GetSDRCommandRequest{}
		if err := m.Request(request); err != nil {
			return err
		}
		response.ReadData = data1[request.OffsetIntoRecord : request.OffsetIntoRecord+request.ByteToRead]
		return response
	})

	res_senReading := &GetSensorReadingResponse{}
	res_senReading.CompletionCode = CommandCompleted
	res_senReading.SensorReading = 0x2a
	res_senReading.ReadingAvail = 0xc0
	res_senReading.Data1 = 0xc0
	res_senReading.Data2 = 0x00
	s.SetHandler(NetworkFunctionSensorEvent, CommandGetSensorReading, func(m *Message) Response {
		return res_senReading
	})

	sdrSensorInfoList, err := client.GetSensorList(reserve.ReservationId)
	assert.Equal(t, nil, err)

	if err == nil {
		assert.Equal(t, 1, len(sdrSensorInfoList))
		if len(sdrSensorInfoList) >= 1 {
			assert.Equal(t, "Fan", sdrSensorInfoList[0].SensorType)
			assert.Equal(t, float64(2646), sdrSensorInfoList[0].Value)
			assert.Equal(t, "Fan 5", sdrSensorInfoList[0].DeviceId)
			assert.Equal(t, "RPM", sdrSensorInfoList[0].BaseUnit)
			assert.Equal(t, true, sdrSensorInfoList[0].Avail)
		}
	}
}

func TestGetSensorStatDesc(t *testing.T) {
	var readingType SDRSensorReadingType = SENSOR_READTYPE_THREADHOLD
	var sensorType SDRSensorType
	var state2 uint8 = 0x20
	var state3 uint8 = 0x00
	var sensorStatDescStr string
	var sensorEve  []string
	var sensorEve2 []string

	sensorStatDescStr, sensorEve = GetSensorStatDesc(readingType, SDRSensorType(0x00), state2, uint8(0))
	assert.Equal(t, sensorStatDescStr, "Upper Non-Recoverable")

	readingType = 0x0b
	sensorType = SDR_SENSOR_TYPECODES_MEMORY
	state2 = 0x7f
	state3 = 0x00
	sensorStatDescStr, sensorEve = GetSensorStatDesc(readingType, sensorType, state2, state3)
	assert.Equal(t,"Non-redundant:Sufficient Resources from Redundant",sensorEve[3])

	readingType = 0x6f
	sensorType = SDR_SENSOR_TYPECODES_CRITICALINTERRUPT
	state2 = 0xfc
	state3 = 0x03
	sensorStatDescStr, sensorEve2 = GetSensorStatDesc(readingType, sensorType, state2, state3)
	assert.Equal(t,"PCI SERR",sensorEve2[3])
	assert.Equal(t,8,len(sensorEve2))

}
