package main

import (
	"golog/log"

	"time"
	"tcpDemo/tcpserver"
	"sync"
	"fmt"
	"net"
	"bufio"
	"io"
)

var (
	quit       = make(chan struct{})
	wg         = &sync.WaitGroup{}
	portString = fmt.Sprintf(":%d", port)
	logger     = log.GetLogger("./logs/app")
)

const (
	port = 37073
)

func main() {

	defer log.Uninit(logger)

	srv, err := tcpserver.New(portString, newEchoAgent, 100)
	if err != nil {
		logger.LError("%s", err)
		return
	}
	err = srv.Start()
	if err != nil {
		logger.LError("%s", err)
		return
	}

	for i := 1; i <= 500; i++ {
		i := i
		wg.Add(1)
		go echoClient(i * 73)
		time.Sleep(time.Millisecond)
	}

	wg.Wait()

	srv.Stop()
	waitSrv := make(chan struct{})
	go func() {
		srv.Wait()
		close(waitSrv)
	}()
	select {
	case <-waitSrv:
	case <-time.After(time.Second * 3):
	}
}

func echoClient(seed int) {
	defer wg.Done()
	conn, err := net.Dial("tcp", fmt.Sprintf("localhost%s", portString))
	if err != nil {
		logger.LError("%s", err)

		return
	}
	defer conn.Close()

	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)

	c := seed
	for c < 10 {
		msg := fmt.Sprintf("%d\n", c)
		writer.Write([]byte(msg))
		writer.Flush()

		line, err := reader.ReadBytes('\n')
		if err == io.EOF {
			logger.LError("%s", err)

			return
		}
		if neterr, ok := err.(net.Error); ok && neterr.Timeout() {
			logger.LError("%s", err)

			return
		}

		if string(line) != msg {
			logger.LFatal("")
		}

		select {
		case <-quit:
			return
		default:
		}

		c++
	}
}

//-----------------------------------------------------------------------------
// our echo agent factory

func newEchoAgent(conn net.Conn, reader *bufio.Reader, writer *bufio.Writer, quit chan struct{}) tcpserver.Agent {
	return &echoAgent{conn, reader, writer, quit}
}

//-----------------------------------------------------------------------------
// our echo agent

type echoAgent struct {
	conn   net.Conn
	reader *bufio.Reader
	writer *bufio.Writer
	quit   chan struct{}
}

func (x *echoAgent) Proceed() error {
	x.conn.SetDeadline(time.Now().Add(time.Second * 3))
	select {
	case <-x.quit:

		return tcpserver.Error(`quit`)
	default:
	}
	line, err := x.reader.ReadBytes('\n')
	if err != nil {
		if err != io.EOF {
			logger.LError("%s", err)

		}
		return err
	}
	_, err = x.writer.Write(line)
	if err != nil {
		if err != io.EOF {
			logger.LError("%s", err)

		}
		return err
	}
	err = x.writer.Flush()
	if err != nil {
		if err != io.EOF {
			logger.LError("%s", err)

		}
		return err
	}

	return nil
}
