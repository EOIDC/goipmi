// +build linux

package ipmi

// #include <stdlib.h>
// #include "ipmi.h"
import "C"

import (
	"errors"
	"unsafe"
	"fmt"
)

// local is another implementation of `transport` interface
// using linux local driver
// open/close/send is implemented
type local struct {
	*Connection
	ctx C.ipmi_ctx
}

func newOpenTransport(c *Connection) transport {
	return &local{Connection: c}
}

func (l *local) open() error {
	rv := C.ipmi_open(&l.ctx)
	if rv == 0 {
		return nil
	}

	return errors.New(fmt.Sprintf("Failed to open local ipmi driver, errno is %d", rv))
}

func (l *local) close() error {
	C.ipmi_close(&l.ctx)
	return nil
}

func (l *local) send(req *Request, resp Response) error {
	//send data to ipmi
	var request C.ipmi_rq
	var response C.ipmi_rsp
	request.netfn = C.uchar(req.NetworkFunction)
	request.lun = C.uchar(0)
	request.cmd = C.uchar(req.Command)
	if req.Data != nil {
		b := messageDataToBytes(req.Data)
		// Go []byte slice to C array
		// The C array is allocated in the C heap using malloc.
		// It is the caller's responsibility to arrange for it to be
		// freed, such as by calling C.free (be sure to include stdlib.h
		// if C.free is needed).
		rData := C.CBytes(b)
		defer C.free(rData)
		request.data = (*C.uchar)(rData)
		request.data_len = C.ushort(len(b))
	}
	rv := C.ipmi_send(&l.ctx, &request, &response)
	if rv != 0 {
		return errors.New(fmt.Sprintf("Faild to write command and recv from local ipmi driver, errno is %d", rv))
	}
	respData := C.GoBytes(unsafe.Pointer(&response.data), response.data_len)
	if ( CompletionCode(respData[0]) != CommandCompleted ) {
		return CompletionCode(respData[0])
	} else {
		return messageDataFromBytes(respData, resp)
	}
}

func (l *local) Console() error {
	return errors.New("Not implement yet.")
}
