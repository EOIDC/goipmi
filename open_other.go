// +build !linux

package ipmi


func newOpenTransport(c *Connection) transport {
	panic("only linux support transport as open")
}
