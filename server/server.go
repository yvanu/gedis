package server

import (
	"context"
	"gedis/gedis"
	"gedis/tool/logger"
	"net"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

type Server struct {
	listener      net.Listener
	waitDone      sync.WaitGroup
	clientCounter int64

	addr    string
	close   int32
	quit    chan os.Signal
	handler gedis.Handler
}

func NewServer(addr string, handler gedis.Handler) *Server {
	return &Server{
		addr:    addr,
		handler: handler,
		quit:    make(chan os.Signal, 1),
	}
}

func (s *Server) Start() error {
	listen, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}
	s.listener = listen
	logger.Infof("监听：%s\n", s.addr)
	go s.accept()
	signal.Notify(s.quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGQUIT)
	<-s.quit
	return nil
}

func (s *Server) accept() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Timeout() {
				logger.Infof("连接超时，5s后重试\n")
				time.Sleep(5 * time.Second)
				continue
			}
			// 这里肯定是listen出问题了
			logger.Warn(err.Error())
			atomic.CompareAndSwapInt32(&s.close, 0, 1)
			s.quit <- syscall.SIGTERM
			break
		}
		go s.handleConn(conn)
	}
}

func (s *Server) handleConn(conn net.Conn) {
	if atomic.LoadInt32(&s.close) == 1 {
		conn.Close()
		return
	}
	logger.Debugf("收到新的连接: %s\n", conn.RemoteAddr().String())
	s.waitDone.Add(1)
	atomic.AddInt64(&s.clientCounter, 1)
	defer func() {
		s.waitDone.Done()
		atomic.AddInt64(&s.clientCounter, -1)
	}()
	s.handler.Handle(context.TODO(), conn)
}

func (s *Server) Close() {
	logger.Info("停止server")
	atomic.CompareAndSwapInt32(&s.close, 0, 1)
	s.listener.Close()
	s.handler.Close()
	s.waitDone.Wait()
}
