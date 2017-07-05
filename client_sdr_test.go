package ipmi

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
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
	r1.deviceId = "fullsensor deviceid"
	r1.Unit = 0x00
	r1.SensorNumber = 0x04
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
		assert.Equal(t, "fullsensor deviceid", r11.DeviceId())
		assert.Equal(t, uint16(65535), nextRecordId1)
	}

	r2 := &SDRCompactSensor{}
	r2.Recordid = 10
	r2.Rtype = SDR_RECORD_TYPE_COMPACT_SENSOR
	r2.SDRVersion = 0x51
	r2.deviceId = "compactsensor deviceId"
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
		assert.Equal(t, "compactsensor deviceId", r22.DeviceId())
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
	res, avail := calFullSensorValue(fs1, 0x11)
	assert.Equal(t, float64(136.0), res)
	assert.Equal(t, true, avail)

	fs1.SetMBExp(1, 0, 0, 0)
	fs1.ReadingType = SENSOR_READTYPE_THREADHOLD
	fs1.Unit = 0x80
	res, avail = calFullSensorValue(fs1, 0xcf)
	assert.Equal(t, float64(-49.0), res)
	assert.Equal(t, true, avail)

	fs1.SetMBExp(2, 0, 0, -2)
	fs1.ReadingType = SENSOR_READTYPE_THREADHOLD
	fs1.Unit = 0x00
	res, avail = calFullSensorValue(fs1, 0xa8)
	assert.Equal(t, float64(3.36), res)

	fs1.SetMBExp(2, 0, 0, -2)
	fs1.ReadingType = SENSOR_READTYPE_THREADHOLD
	fs1.Unit = 0x00
	res, avail = calFullSensorValue(fs1, 0xa8)
	assert.Equal(t, float64(3.36), res)

	assert.Equal(t, true, avail)
}
func TestCalCompactSensorValue(t *testing.T) {
	cs1 := &SDRCompactSensor{}
	cs1.ReadingType = SENSOR_READTYPE_SENSORSPECIF
	cs1.Unit = 0xc0
	res, avail := calCompactSensorValue(cs1, 0x11)
	assert.Equal(t, float64(17), res)
	assert.Equal(t, true, avail)
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
		}
	})

	err = client.Open()
	assert.NoError(t, err)

	if SensorReading, err := client.getSensorReading(0x04); err == nil {
		assert.Equal(t, uint8(56), SensorReading)
	}
}
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
	r1.deviceId = "fullsensor deviceid"
	r1.Unit = 0x00
	r1.SensorNumber = 0x04
	r1.BaseUnit = 0x12
	r1.SetMBExp(63, 0, 0, 0)
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
	sdrSensorInfoList, err := client.GetSensorList(reserve.ReservationId)
	assert.Equal(t, nil, err)
	assert.Equal(t, 1, len(sdrSensorInfoList))
	assert.Equal(t, "fullsensor deviceid", sdrSensorInfoList[0].DeviceId)
	assert.Equal(t, "RPM", sdrSensorInfoList[0].BaseUnit)
}
