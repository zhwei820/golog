package main

import (
	"golog/log"
	"fmt"
	"net"
	"os"
)

var host = "log.wxdesk.com"
var port = "22002"

func udpSend(s string) {
	addr, err := net.ResolveUDPAddr("udp", host+":"+port)
	if err != nil {
		logger.LError("Can't resolve address: %s", err)
		os.Exit(1)
	}
	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		logger.LError("%s", err)
	}
	defer conn.Close()
	conn.Write([]byte(s))
	logger.LInfo("<%s>\n", conn.RemoteAddr())
}
func udpMessage(snbid string, key string, body_bytes string) {
	header_bytes := "{\"key\": \"" + key + "\"}" //ujson.dumps({"key": key}, ensure_ascii=False).encode()
	header_length := len(header_bytes)
	body_length := len(body_bytes)
	total_length := header_length + body_length + 12
	s := fmt.Sprintf("sicent%06x%06x%s%06x%s", total_length, header_length, header_bytes, body_length, body_bytes)
	logger.LInfo(s)

	udpSend(s)
}

var logger = log.GetLogger("./logs/app")

func main() {
	defer log.Uninit(logger)
	for ii := 0; ii < 100; ii++ {
		udpMessage("huqiuxia", "testudplog", "{\"snbid\":\"huqiuxia\",\"test\":\"ok\"}")
	}
	log.SetLevel(log.LvTRACE)
	logger.LInfo("app exit")
}
