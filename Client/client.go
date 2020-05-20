package Client

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"git.garena.com/changyou.liang/entry/msgproto"
	"github.com/golang/protobuf/proto"
	"io/ioutil"
)

type ClinetInterface interface {
	// login success， return cookie
	Login(passwd string) (string, error)
	WriteSecureMessage(key, value string) error
	ReadSecureMessage(key string) (string, error)

	Close() error
}

func NewClient(user, addr string, clientPem, clientKey string, serverPem string) (*Client, error) {

	cert, err := tls.LoadX509KeyPair(clientPem, clientKey)
	if err != nil {
		return nil, err
	}
	certBytes, err := ioutil.ReadFile(clientPem)
	if err != nil {
		return nil, err
	}
	clientCertPool := x509.NewCertPool()
	ok := clientCertPool.AppendCertsFromPEM(certBytes)
	if !ok {
		return nil, errors.New("failed to parse root certificate")
	}
	conf := &tls.Config{
		RootCAs:            clientCertPool,
		Certificates:       []tls.Certificate{cert},
		InsecureSkipVerify: true,
	}
	conn, err := tls.Dial("tcp", addr, conf)
	if err != nil {
		return nil, err
	}
	return &Client{
		Conn: conn,
		User: user,
	}, nil
}

type Client struct {
	Conn   *tls.Conn
	User   string
	Cookie string
}

// sent server user passwd return cookie
func (c *Client) Login(passwd string) (string, error) {
	msg := msgproto.Msg{
		Cmd:   1,
		User:  c.User,
		Value: passwd,
	}
	buf, err := proto.Marshal(&msg)
	if err != nil {
		return "", err
	}
	_, err = c.write(buf)
	if err != nil {
		return "", err
	}

	ck, err := c.read()
	if err != nil {
		return "", err
	}
	rep := new(msgproto.Reply)
	err = proto.Unmarshal(ck, rep)
	if err != nil {
		return "", err
	}
	// TODO rep.code

	if rep.Code != 0 {
		return rep.Msg, errors.New(rep.Msg)
	}
	c.SetCookie(rep.Msg)
	return rep.Msg, nil
}

func (c *Client) WriteSecureMessage(key, value string) error {
	msg := msgproto.Msg{
		Cmd:    2,
		User:   c.User,
		Key:    key,
		Value:  value,
		Cookie: c.Cookie,
	}

	buf, err := proto.Marshal(&msg)
	if err != nil {
		return err
	}
	_, err = c.write(buf)
	if err != nil {
		fmt.Println("err")
		return err
	}

	ck, err := c.read()
	if err != nil {
		return err
	}
	rep := new(msgproto.Reply)
	err = proto.Unmarshal(ck, rep)
	if err != nil {
		return err
	}

	if rep.Code != 0 {
		return errors.New(rep.Msg)
	}
	return nil
}

func (c *Client) ReadSecureMessage(key string) (string, error) {
	msg := msgproto.Msg{
		Cmd:    3,
		User:   c.User,
		Key:    key,
		Cookie: c.Cookie,
	}

	// 抽取出来
	buf, err := proto.Marshal(&msg)
	if err != nil {
		return "", err
	}
	_, err = c.write(buf)
	if err != nil {
		return "", err
	}

	ck, err := c.read()
	if err != nil {
		return "", err
	}
	rep := new(msgproto.Reply)
	err = proto.Unmarshal(ck, rep)
	if err != nil {
		return "", err
	}
	return rep.Msg, err
}

func (c *Client) SetCookie(cookie string) {
	c.Cookie = cookie
}

func (c *Client) write(data []byte) (int, error) {
	return c.Conn.Write(data)
}

func (c *Client) read() ([]byte, error) {
	buf := make([]byte, 2550)
	l, err := c.Conn.Read(buf)
	if err != nil {
		return nil, err
	} else {
		return buf[:l], nil
	}
	//return ioutil.ReadAll(c.Conn)
}
