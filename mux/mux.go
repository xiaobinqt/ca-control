package mux

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

type MuxHandle interface {
	Detect(header []byte) bool
	Handle(net.Conn) error
}

type Mux struct {
	l              net.Listener
	conns          chan net.Conn
	muxs           map[string]MuxHandle
	headerReadFull int
	tlsConfig      *tls.Config
}

func NewMux(tlsConfig *tls.Config) *Mux {
	return &Mux{
		conns:          make(chan net.Conn),
		muxs:           make(map[string]MuxHandle),
		tlsConfig:      tlsConfig,
		headerReadFull: 5, // 默认读取 5 个字节
	}
}

func (s *Mux) SetHeaderReadFull(n int) {
	s.headerReadFull = n
}

func (sp *Mux) RegisterHandler(name string, d MuxHandle) {
	sp.muxs[name] = d
}

func (sp *Mux) ListenAndServe(addr string) error {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	sp.l = l
	err = sp.Serve(l)
	l.Close()
	return err
}

func (sp *Mux) Serve(l net.Listener) error {
	sp.l = l
	var tempDelay time.Duration // how long to sleep on accept failure
	for {
		conn, e := l.Accept()
		if e != nil {
			if _, ok := e.(*net.OpError); ok {
				logrus.Errorf("Mux: network op error: %v", e.Error())
				return e
			}

			if ne, ok := e.(net.Error); ok && ne.Temporary() {
				if tempDelay == 0 {
					tempDelay = 5 * time.Millisecond
				} else {
					tempDelay *= 2
				}
				if max := 1 * time.Second; tempDelay > max {
					tempDelay = max
				}
				logrus.Warnf("Mux: Accept error: %v; retrying in %v", e, tempDelay)
				time.Sleep(tempDelay)
				continue
			}
			return e
		}
		tempDelay = 0
		go sp.dispatch(conn)
	}
}

func (sp *Mux) dispatch(conn net.Conn) {
	if tconn, ok := conn.(*net.TCPConn); ok {
		tconn.SetKeepAlive(true)
		tconn.SetKeepAlivePeriod(time.Second * 30)
	}

	buf := new(bytes.Buffer)
	header := make([]byte, sp.headerReadFull)
	_, err := io.ReadFull(io.TeeReader(conn, buf), header)
	if err != nil {
		conn.Close()
		logrus.Errorf("Failed to read protocol header: %v", err)
		return
	}

	logrus.Debugf("dispatch read header: %s", string(header))

	bc := &bufConn{
		Conn: conn,
		r:    io.MultiReader(buf, conn),
	}

	var (
		name  string
		mux   MuxHandle
		found bool
	)

	for name, mux = range sp.muxs {
		if mux.Detect(header) {
			found = true
			break
		}
	}

	if !found {
		logrus.Errorf("Unknown protocol header: %s", string(header))
		conn.Close()
		return
	}

	if err = mux.Handle(bc); err != nil {
		logrus.Errorf("Failed to handle %s protocol: %v", name, err)
	}
}

// Accept implement net.Listener
func (sp *Mux) Accept() (net.Conn, error) {
	c, ok := <-sp.conns
	if ok {
		return c, nil
	}

	return nil, fmt.Errorf("listener closed")
}

func (sp *Mux) Close() error {
	sp.l.Close()
	close(sp.conns)
	return nil
}

func (sp *Mux) Addr() net.Addr {
	return sp.l.Addr()
}

type bufConn struct {
	net.Conn
	r io.Reader
	sync.Mutex
}

func (bc *bufConn) Read(b []byte) (int, error) {
	bc.Lock()
	defer bc.Unlock()
	return bc.r.Read(b)
}
