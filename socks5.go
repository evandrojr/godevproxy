// ...existing code...
package main

import (
	"fmt"
	"io"
	"net"
	"strings"
)

type SOCKS5Server struct {
	port string
}

func NewSOCKS5Server(port string) *SOCKS5Server {
	return &SOCKS5Server{port: port}
}

func (s *SOCKS5Server) Start() error {
	listener, err := net.Listen("tcp", ":"+s.port)
	if err != nil {
		return fmt.Errorf("erro ao iniciar servidor na porta %s: %v", s.port, err)
	}
	defer listener.Close()

   logf("🚀 GoDevProxy SOCKS%d iniciado na porta %s\n", SocksVersion, s.port)
   logf("🔑 Usuário: admin\n")
   logf("🔑 Senha:   123\n")
   logf("📋 Para testar: configure proxy SOCKS5 como localhost:%s\n", s.port)
   logf("🔧 Exemplo curl: curl --socks5 %s:%s@localhost:%s https://g1.globo.com\n",
	   func() string { if authUser != "" { return authUser } else { return "admin" } }(),
	   func() string { if authPass != "" { return authPass } else { return "123" } }(),
	   s.port)
   logf("📝 Logs de conexões do GoDevProxy aparecerão abaixo:\n")
   logf("%s\n", strings.Repeat("=", 60))

   for {
	   conn, err := listener.Accept()
	   if err != nil {
		   logf("❌ Erro ao aceitar conexão: %v\n", err)
		   continue
	   }
	   logf("🔗 Nova conexão de: %s\n", conn.RemoteAddr())
	   go s.handleConnection(conn)
   }
}

func (s *SOCKS5Server) handleConnection(conn net.Conn) {
	defer conn.Close()
	clientAddr := conn.RemoteAddr().String()
   if !HandleSocks5Auth(conn, clientAddr) {
	   return
   }
   s.handleRequest(conn, clientAddr)
}

func (s *SOCKS5Server) handleRequest(conn net.Conn, clientAddr string) {
	buf := make([]byte, 256)
   n, err := conn.Read(buf)
   if err != nil {
	   logf("❌ [%s] Erro ao ler request: %v\n", clientAddr, err)
	   return
   }
   if n < 7 || buf[0] != SocksVersion || buf[1] != CONNECT {
	   logf("❌ [%s] Requisição inválida\n", clientAddr)
	   s.sendErrorResponse(conn)
	   return
   }
	var destHost string
	var destPort int
	var destType string
   switch buf[3] {
   case IPV4:
	   destType = "IPV4"
	   if n < 10 {
		   logf("❌ [%s] Requisição IPv4 malformada\n", clientAddr)
		   s.sendErrorResponse(conn)
		   return
	   }
	   destHost = fmt.Sprintf("%d.%d.%d.%d", buf[4], buf[5], buf[6], buf[7])
	   destPort = int(buf[8])<<8 + int(buf[9])
   case DOMAIN:
	   destType = "DOMAIN"
	   if n < 7 {
		   logf("❌ [%s] Requisição de domínio malformada\n", clientAddr)
		   s.sendErrorResponse(conn)
		   return
	   }
	   domainLen := int(buf[4])
	   if n < 7+domainLen {
		   logf("❌ [%s] Requisição de domínio incompleta\n", clientAddr)
		   s.sendErrorResponse(conn)
		   return
	   }
	   destHost = string(buf[5 : 5+domainLen])
	   destPort = int(buf[5+domainLen])<<8 + int(buf[6+domainLen])
   default:
	   destType = fmt.Sprintf("0x%02x", buf[3])
	   logf("❌ [%s] Tipo de endereço não suportado: %02x\n", clientAddr, buf[3])
	   s.sendErrorResponse(conn)
	   return
	}
	destAddr := fmt.Sprintf("%s:%d", destHost, destPort)
	// LOG DETALHADO DA REQUISIÇÃO SOCKS5
   logf("🧦 [%s] SOCKS5 %s para %s (%s)\n", clientAddr, destType, destAddr, destHost)
   logf("🎯 [%s] Conectando para: %s\n", clientAddr, destAddr)
   destConn, err := net.Dial("tcp", destAddr)
   if err != nil {
	   logf("❌ [%s] Falha ao conectar em %s: %v\n", clientAddr, destAddr, err)
	   s.sendErrorResponse(conn)
	   return
   }
   defer destConn.Close()
   logf("✅ [%s] Conexão estabelecida com %s\n", clientAddr, destAddr)
   response := []byte{
		SocksVersion, SUCCESS, 0x00,
		IPV4, 0, 0, 0, 0,
		0, 0,
	}
   _, err = conn.Write(response)
   if err != nil {
	   logf("❌ [%s] Erro ao enviar success response: %v\n", clientAddr, err)
	   return
   }
   logf("🔄 [%s] Iniciando relay de dados para %s\n", clientAddr, destAddr)
   go s.relay(conn, destConn, fmt.Sprintf("%s->%s", clientAddr, destAddr))
   s.relay(destConn, conn, fmt.Sprintf("%s<-%s", clientAddr, destAddr))
   logf("🔚 [%s] Conexão finalizada\n", clientAddr)
}

func (s *SOCKS5Server) sendErrorResponse(conn net.Conn) {
	response := []byte{
		SocksVersion, FAILURE, 0x00,
		IPV4, 0, 0, 0, 0,
		0, 0,
	}
   conn.Write(response)
}

func (s *SOCKS5Server) relay(src, dst net.Conn, direction string) {
	defer dst.Close()
	defer src.Close()
	buf := make([]byte, 4096)
	totalBytes := 0
   for {
	   n, err := src.Read(buf)
	   if err != nil {
		   if err != io.EOF {
			   logf("🔌 [%s] Conexão encerrada: %v\n", direction, err)
		   }
		   break
	   }
	   _, err = dst.Write(buf[:n])
	   if err != nil {
		   logf("❌ [%s] Erro ao escrever: %v\n", direction, err)
		   break
	   }
	   totalBytes += n
	   if totalBytes > 0 && totalBytes%1024 == 0 {
		   logf("📊 [%s] Transferidos %d bytes", direction, totalBytes)
	   }
   }
   logf("📈 [%s] Total transferido: %d bytes", direction, totalBytes)
}
