package Server

import (
	"crypto/md5"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"errors"
	"fmt"
	"git.garena.com/changyou.liang/entry/conf"
	"git.garena.com/changyou.liang/entry/db"
	"git.garena.com/changyou.liang/entry/msgproto"
	"github.com/gomodule/redigo/redis"
	"github.com/jacksonyoudi/ratelimit"
	"google.golang.org/protobuf/proto"
	"io/ioutil"
	"log"
	"net"
	"time"
)

type Service interface {
	Start()
	Close()
	Handle(*ConnNode)
}

type TcpService struct {
	// bucket map
	Listner net.Listener
	DBCfg   struct {
		Dsn    string
		Driver string
	}
	RedisCfg struct {
		Address string
	}

	Tls struct {
		Server struct {
			Pem string
			key string
		}
		Client struct {
			Pem string
			Key string
		}
	}
	RedisCache db.Cache

	ReadRateBucketMap  *BucketMap
	WriteRateBucketMap *BucketMap
}

func (s TcpService) Start() {
	log.Println("service is Listner")

	for {
		conn, err := s.Listner.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		go s.Handle(&ConnNode{
			Conn: conn,
		})
	}
}

func (s TcpService) Close() {
	// clean
	s.Listner.Close()

}

func NewTcpService(cfg *conf.ServiceCfg) (*TcpService, error) {
	cert, err := tls.LoadX509KeyPair(cfg.Tls.Server.Pem, cfg.Tls.Server.Key)
	if err != nil {
		return nil, err
	}
	certBytes, err := ioutil.ReadFile(cfg.Tls.Client.Pem)
	if err != nil {
		return nil, err
	}
	clientCertPool := x509.NewCertPool()
	ok := clientCertPool.AppendCertsFromPEM(certBytes)
	if !ok {
		return nil, errors.New("failed to parse root certificate")
	}
	config := &tls.Config{
		Certificates: []tls.Certificate{cert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    clientCertPool,
	}

	listener, err := tls.Listen("tcp", cfg.Address, config)
	if err != nil {
		return nil, err
	}

	// init redis
	c, err := redis.Dial("tcp", cfg.Redis.Address)
	if err != nil {
		return nil, err
	}

	return &TcpService{
		RedisCache: &db.RedisCache{
			Conn: c,
		},
		Listner: listener,
		DBCfg: struct {
			Dsn    string
			Driver string
		}{Dsn: cfg.MySql.Dsn, Driver: cfg.MySql.Driver},
		RedisCfg: struct {
			Address string
		}{Address: cfg.Redis.Address},

		Tls: struct {
			Server struct {
				Pem string
				key string
			}
			Client struct {
				Pem string
				Key string
			}
		}{Server: struct {
			Pem string
			key string
		}{Pem: cfg.Tls.Server.Pem, key: cfg.Tls.Server.Key},
			Client: struct {
				Pem string
				Key string
			}{Pem: cfg.Tls.Client.Pem, Key: cfg.Tls.Client.Key}},

		ReadRateBucketMap:  new(BucketMap),
		WriteRateBucketMap: new(BucketMap),
	}, nil
}

func (s TcpService) Handle(node *ConnNode) {
	defer node.close()
	buf, err := node.read()
	if err != nil {
		log.Println(err)
		return
	}

	msg := new(msgproto.Msg)
	err = proto.Unmarshal(buf, msg)
	if err != nil {
		log.Println(err)
		return
	}
	err = s.filter(node, msg)

}

func (s TcpService) filter(node *ConnNode, msg *msgproto.Msg) error {
	reply := msgproto.Reply{}
	reply.Code = -1
	var errR error

	if len(msg.GetUser()) == 0 {
		reply.Msg = "no user"
		errR = errors.New("no user")
	}

	switch msg.GetCmd() {
	// login
	case 1:
		sid, rq, wq, err := s.login(msg)
		if err != nil {
			reply.Msg = "login err"
			errR = err
		} else {
			err := s.setSession(msg.GetUser(), sid)
			errR = err
			if err != nil {
				reply.Msg = "set session err"
			} else {
				reply.Code = 0
				reply.Msg = sid

				rb := ratelimit.NewBucketWithRate(float64(rq), int64(rq))
				s.ReadRateBucketMap.set(msg.GetUser(), rb)

				wb := ratelimit.NewBucketWithRate(float64(wq), int64(wq))
				s.WriteRateBucketMap.set(msg.GetUser(), wb)
			}
		}
		// write
	case 2:

		err := s.checkSession(msg.GetCookie())
		if err != nil {
			reply.Msg = err.Error()
			errR = err
		} else {
			ok := s.rateLimit(2, msg.GetUser())
			if ok {
				err := s.setData(msg.GetUser(), msg.GetKey(), msg.GetValue())
				errR = err
				if err != nil {
					fmt.Println(err)
					reply.Msg = err.Error()
				} else {
					reply.Code = 0
					reply.Msg = "set data ok"
				}
			} else {
				reply.Code = -1
				reply.Msg = "request write rate limit"
			}
		}
		// read
	case 3:
		err := s.checkSession(msg.GetCookie())
		if err != nil {
			reply.Msg = err.Error()
			errR = err
		} else {
			ok := s.rateLimit(3, msg.GetUser())
			if ok {
				data, err := s.getData(msg.GetUser(), msg.GetKey())
				errR = err
				if err != nil {
					reply.Msg = err.Error()
				} else {
					reply.Code = 0
					reply.Msg = data
				}
			} else {
				reply.Code = -1
				reply.Msg = "request read rate limit"
			}
		}
	default:
		errR = errors.New("illegal request")
	}

	buf, err := proto.Marshal(&reply)
	if err != nil {
		log.Println(err)
	} else {
		_, err := node.write(buf)
		errR = err
	}
	return errR

}

func (s *TcpService) login(msg *msgproto.Msg) (string, int, int, error) {
	rq, wq, err := db.NewUserQuery(msg.GetUser(), msg.GetValue(), s.DBCfg.Driver, s.DBCfg.Dsn)
	if err != nil {
		return "", rq, wq, err
	}
	sessionKey := msg.GetUser() + time.Now().Format(time.RFC3339Nano)
	m := md5.New()
	m.Write([]byte(sessionKey))
	sessionId := hex.EncodeToString(m.Sum(nil))
	return sessionId, rq, wq, nil
}

func (s *TcpService) setSession(user string, sessionId string) error {
	return s.RedisCache.SetSession(user, sessionId)
}

func (s *TcpService) getSession(user string) (string, error) {
	return s.RedisCache.GetSession(user)
}

func (s *TcpService) setData(user string, key string, value string) error {
	return s.RedisCache.SetData(user, key, value)
}

func (s *TcpService) getData(user string, key string) (string, error) {
	return s.RedisCache.GetData(user, key)
}

func (s *TcpService) checkSession(sesssion string) error {
	if len(sesssion) != 32 {
		return errors.New("session is error")
	}
	_, err := s.RedisCache.GetSession(sesssion)
	return err
}

func (s *TcpService) rateLimit(cmd int, user string) bool {
	switch cmd {
	// write
	case 2:
		wb, ok := s.WriteRateBucketMap.get(user)
		if ok {
			j := wb.TakeAvailable(1)
			if j <= 0 {
				return false
			}
			return ok
		}
		return false

	case 3:
		rb, ok := s.ReadRateBucketMap.get(user)
		if ok {
			j := rb.TakeAvailable(1)
			if j <= 0 {
				return false
			}
			return ok
		}
		return false
	default:
		return false
	}
}
