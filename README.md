# goipmi

Native IPMI implementation and `ipmitool` wrapper in Go.

Package goipmi provides utilities to access to i and pass around stack traces.

对使用从ipmi协议进行通信的服务器采集数据提供了底层的接口， 主要是dsn.go对外提供链接服务器接口，client_sdr.go 对外提供采集数据接口，

## Usage

#### type SdrSensorInfo

```go
type SdrSensorInfo struct {
	SensorType  string
	BaseUnit    string
	Value       float64
	DeviceId    string
	StatusDesc  string
	SensorEvent []string
	Avail       bool
}
```

`SdrSensorInfo`定义用户所关心采集数据的结构:

- `SensorType`所采集类型，如"Temperature", "Voltage", "Current", "Fan"
- `BaseUnit`数据的单位，如"degrees C", "Volts", "Amps", "Watts","Joules"
- `Value`所采集数据类型的值(float64)
- `DeviceId`描述所采集Sensor，如“PSU FAULT”，“BIOS”，“Watchdog”等
- `StatusDesc`当sensor类型是Threshold时，该字段表示sensor的状态描述信息
- `SensorEvent`当sensor类型是Sensor-specific或generic时, 该字段表示envent的列表
- `Avail`描述所采集的数据值是否可用，false为unavailable，true为available

#### GetSensorList

```go
func GetSensorList(reservationID uint16) ([]SdrSensorInfo, error) 
```
- reservationID是采集前获取的初始值，用来迭代RecordId直到RecordId等于Oxffff，将获取的信息保存到数组SdrSensorInfo中

#### GetSDR

```go
func GetSDR(reservationID uint16, recordID uint16) (sdr *sDRRecordAndValue, next uint16, err error)
```

执行ipmi命令   ‘Get SDR’commands,reservationID针对一次查询需求生成的ID标识，recordID是查询某一个recordID值， 三次‘get SDR’ Request请求，返回获取到的sensor信息

返回下次迭代查询的recordID值

#### getSensorReading

```go
func getSensorReading(sensorNum uint8) (*GetSensorReadingResponse, error) 
```

执行 ipmi ‘Get Sensor Reading’ Command，根据sensorNum请求获取相应，response返回sensorReading值，可用不可用及状体或Event列表

#### GetSensorStatDesc

```go
func GetSensorStatDesc(readingType SDRSensorReadingType, sensorType SDRSensorType, state2 uint8, state3 uint8) (string, []string) 
```

根据‘get SDR’ command获取到的 readingType、sensorType值，‘get Sensor Reading’ command 获取的state2（既data1）和state3（既data2），返回
当前sensor的状态string和Event列表[]string, 

## License

This project is available under the [Apache 2.0](./LICENSE) license.
