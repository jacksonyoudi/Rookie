package Server

import (
	"net"
)

// only read write
type Connect interface {
	// read data
	read() ([]byte, error)
	// write
	write([]byte) (int, error)
	close() error
}
type ConnNode struct {
	Conn net.Conn
}

func (c *ConnNode) write(buf []byte) (int, error) {
	return c.Conn.Write(buf)
}

// TODO
func (c *ConnNode) read() ([]byte, error) {
	buf := make([]byte, 2550)
	l, err := c.Conn.Read(buf)
	if err != nil {
		return nil, err
	} else {
		return buf[:l], nil
	}
}

func (c *ConnNode) close() error {
	return c.Conn.Close()
}
