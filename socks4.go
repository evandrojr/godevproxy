package main

import (
	"fmt"
	"io"
	"net"
)

type SOCKS4Server struct {
	port string
}

func NewSOCKS4Server(port string) *SOCKS4Server {
	return &SOCKS4Server{port: port}
}

func (s *SOCKS4Server) Start() error {
	listener, err := net.Listen("tcp", ":"+s.port)
	if err != nil {
		return fmt.Errorf("erro ao iniciar servidor SOCKS4 na porta %s: %v", s.port, err)
	}
	defer listener.Close()

	fmt.Printf("🚀 Servidor SOCKS4 iniciado na porta %s\n", s.port)
	fmt.Printf("🔧 Exemplo curl: curl --socks4 localhost:%s https://g1.globo.com\n", s.port)
	fmt.Println("📝 Logs de conexões aparecerão abaixo:")
	fmt.Println("============================================================")

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("❌ Erro ao aceitar conexão: %v\n", err)
			continue
		}
		go handleSOCKS4Conn(conn)
	}
}

func handleSOCKS4Conn(conn net.Conn) {
	defer conn.Close()
	buf := make([]byte, 512)
	n, err := conn.Read(buf)
	if err != nil || n < 9 {
		return
	}
	if buf[0] != 0x04 || buf[1] != 0x01 {
		return // Apenas CONNECT suportado
	}
	port := int(buf[2])<<8 | int(buf[3])
	host := fmt.Sprintf("%d.%d.%d.%d", buf[4], buf[5], buf[6], buf[7])
	userEnd := 8
	for ; userEnd < n && buf[userEnd] != 0; userEnd++ {}
	// Ignora USERID
	destAddr := fmt.Sprintf("%s:%d", host, port)
	fmt.Printf("🧦 [SOCKS4] Conectando para %s\n", destAddr)
	destConn, err := net.Dial("tcp", destAddr)
	if err != nil {
		conn.Write([]byte{0x00, 0x5b, 0, 0, 0, 0, 0, 0}) // Falha
		fmt.Printf("❌ [SOCKS4] Falha ao conectar em %s: %v\n", destAddr, err)
		return
	}
	defer destConn.Close()
	conn.Write([]byte{0x00, 0x5a, 0, 0, 0, 0, 0, 0}) // Sucesso
	fmt.Printf("✅ [SOCKS4] Conexão estabelecida com %s\n", destAddr)
	go io.Copy(destConn, conn)
	io.Copy(conn, destConn)
	fmt.Printf("🔚 [SOCKS4] Conexão finalizada para %s\n", destAddr)
}
