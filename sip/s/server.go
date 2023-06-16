package sip

import (
	"fmt"
	"net"
	"net/http"
	"sync"

	"github.com/xuanxuan000/sipserver/utils"
)

var (
	bufferSize uint16 = 65535 - 20 - 8 // IPv4 max size - IPv4 Header size - UDP Header size
)

// RequestHandler RequestHandler
type RequestHandler func(req *Request, tx *Transaction)

// Server Server
type Server struct {
	udpaddr net.Addr
	conn    Connection

	txs *transacionts

	hmu             *sync.RWMutex
	requestHandlers map[RequestMethod]RequestHandler

	port *Port
	host net.IP
}

// NewServer NewServer
func NewServer() *Server {
	activeTX = &transacionts{txs: map[string]*Transaction{}, rwm: &sync.RWMutex{}}
	srv := &Server{hmu: &sync.RWMutex{},
		txs:             activeTX,
		requestHandlers: map[RequestMethod]RequestHandler{}}
	return srv
}

//	func (s *Server) newTX(key string) *Transaction {
//		return s.txs.newTX(key, s.conn)
//	}
func (s *Server) getTX(key string) *Transaction {
	return s.txs.getTX(key)
}
func (s *Server) mustTX(key string) *Transaction {
	tx := s.txs.getTX(key)
	if tx == nil {
		tx = s.txs.newTX(key, s.conn)
	}
	return tx
}

// ListenUDPServer ListenUDPServer
func (s *Server) ListenUDPServer(addr string) {
	udpaddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		utils.Fatal("net.ResolveUDPAddr err", err, addr)
	}
	s.port = NewPort(udpaddr.Port)
	s.host, err = utils.ResolveSelfIP()
	if err != nil {
		utils.Fatal("net.ListenUDP resolveip err", err, addr)
	}
	udp, err := net.ListenUDP("udp", udpaddr)
	if err != nil {
		utils.Fatal("net.ListenUDP err", err, addr)
	}
	s.conn = newUDPConnection(udp)
	var (
		raddr net.Addr
		num   int
	)
	buf := make([]byte, bufferSize)
	parser := newParser()
	defer parser.stop()
	go s.handlerListen(parser.out)
	for {
		num, raddr, err = s.conn.ReadFrom(buf)
		if err != nil {
			utils.Errorln("udp.ReadFromUDP err", err)
			continue
		}
		parser.in <- newPacket(append([]byte{}, buf[:num]...), raddr)
	}
}

// RegistHandler RegistHandler
func (s *Server) RegistHandler(method RequestMethod, handler RequestHandler) {
	s.hmu.Lock()
	s.requestHandlers[method] = handler
	s.hmu.Unlock()
}
func (s *Server) handlerListen(msgs chan Message) {
	var msg Message
	for {
		msg = <-msgs
		switch tmsg := msg.(type) {
		case *Request:
			req := tmsg
			req.SetDestination(s.udpaddr)
			s.handlerRequest(req)
		case *Response:
			resp := tmsg
			resp.SetDestination(s.udpaddr)
			s.handlerResponse(resp)
		default:
			utils.Errorln("undefind msg type,", tmsg, msg.String())
		}
	}
}
func (s *Server) handlerRequest(msg *Request) {
	tx := s.mustTX(getTXKey(msg))
	utils.Traceln("receive request from:", msg.Source(), ",method:", msg.Method(), "txKey:", tx.key, "message: \n", msg.String())
	s.hmu.RLock()
	handler, ok := s.requestHandlers[msg.Method()]
	s.hmu.RUnlock()
	if !ok {
		utils.Errorln("not found handler func,requestMethod:", msg.Method(), msg.String())
		go handlerMethodNotAllowed(msg, tx)
		return
	}

	go handler(msg, tx)
}

func (s *Server) handlerResponse(msg *Response) {
	tx := s.getTX(getTXKey(msg))
	if tx == nil {
		utils.Infoln("not found tx. receive response from:", msg.Source(), "message: \n", msg.String())
	} else {
		utils.Traceln("receive response from:", msg.Source(), "txKey:", tx.key, "message: \n", msg.String())
		tx.receiveResponse(msg)
	}
}

// Request Request
func (s *Server) Request(req *Request) (*Transaction, error) {
	viaHop, ok := req.ViaHop()
	if !ok {
		return nil, fmt.Errorf("missing required 'Via' header")
	}
	viaHop.Host = s.host.String()
	viaHop.Port = s.port
	if viaHop.Params == nil {
		viaHop.Params = NewParams().Add("branch", String{Str: GenerateBranch()})
	}
	if !viaHop.Params.Has("rport") {
		viaHop.Params.Add("rport", nil)
	}

	tx := s.mustTX(getTXKey(req))
	return tx, tx.Request(req)
}

func handlerMethodNotAllowed(req *Request, tx *Transaction) {
	resp := NewResponseFromRequest("", req, http.StatusMethodNotAllowed, http.StatusText(http.StatusMethodNotAllowed), []byte{})
	tx.Respond(resp)
}
