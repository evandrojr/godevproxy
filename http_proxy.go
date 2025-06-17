
package main

import (
	"encoding/base64"
	"fmt"
	"io"
	"net"
	"strings"
	"time"
)

func StartHTTPProxy(port string) error {
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return fmt.Errorf("erro ao iniciar proxy HTTP na porta %s: %v", port, err)
	}
	defer listener.Close()

   logf("🚀 GoDevProxy HTTP iniciado na porta %s\n", port)
   if authUser != "" && authPass != "" {
	   logf("🔑 Usuário: %s\n", authUser)
	   logf("🔑 Senha:   %s\n", authPass)
	   logf("🔧 Exemplo curl: curl -x http://%s:%s@localhost:%s https://g1.globo.com\n", authUser, authPass, port)
   } else {
	   logf("🔓 Sem autenticação (livre)\n")
	   logf("🔧 Exemplo curl: curl -x http://localhost:%s https://g1.globo.com\n", port)
   }
   logf("📝 Logs de conexões do GoDevProxy aparecerão abaixo:\n")
   logf("%s\n", strings.Repeat("=", 60))

   for {
	   conn, err := listener.Accept()
	   if err != nil {
		   logf("❌ Erro ao aceitar conexão: %v\n", err)
		   continue
	   }
	   go handleHTTPProxyConn(conn)
   }
}

func handleHTTPProxyConn(client net.Conn) {
	defer client.Close()
   remoteAddr := client.RemoteAddr().String()
	buf := make([]byte, 4096)
	n, err := client.Read(buf)
	if err != nil {
		return
	}
	req := string(buf[:n])
	lines := strings.Split(req, "\r\n")
	var authOK bool
	reason := ""
	receivedUser := ""
	receivedPass := ""
	if authUser == "" && authPass == "" {
		authOK = true // Sem autenticação
	} else {
		headerFound := false
		for _, line := range lines {
			if strings.HasPrefix(strings.ToLower(line), "proxy-authorization: basic ") {
				headerFound = true
				b64 := strings.TrimSpace(line[len("proxy-authorization: basic "):])
				decoded, err := base64.StdEncoding.DecodeString(b64)
				if err != nil {
					reason = "Base64 inválido no header Proxy-Authorization"
					break
				}
				parts := strings.SplitN(string(decoded), ":", 2)
				if len(parts) != 2 {
					reason = "Formato inválido no header Proxy-Authorization (esperado user:pass)"
					break
				}
				receivedUser = parts[0]
				receivedPass = parts[1]
				if CheckUserPass(receivedUser, receivedPass) {
					authOK = true
					break
				} else {
					reason = "Usuário ou senha incorretos"
					break
				}
			}
		}
		if !headerFound {
			// Só define reason se autenticação for realmente exigida
			// (não logar "Header Proxy-Authorization ausente" se não for necessário)
			// reason = "Header Proxy-Authorization ausente"
		}
	}
   if !authOK {
	   resp := "HTTP/1.1 407 Proxy Authentication Required\r\nProxy-Authenticate: Basic realm=\"Proxy\"\r\n\r\n"
	   client.Write([]byte(resp))
	   logf("❌ [%s] Falha de autenticação HTTP Proxy: %s\n", remoteAddr, reason)
	   logf("   → Usuário recebido: '%s' | Senha recebida: '%s'\n", receivedUser, receivedPass)
	   return
   } else if authUser != "" && authPass != "" {
	   logf("✅ [%s] Autenticação HTTP Proxy bem-sucedida: Usuário='%s' | Senha='%s'\n", remoteAddr, receivedUser, receivedPass)
   }

	// Detectar se é CONNECT (HTTPS) ou requisição HTTP normal
   if strings.HasPrefix(req, "CONNECT ") {
	   // Exemplo: CONNECT g1.globo.com:443 HTTP/1.1
	   parts := strings.SplitN(req, " ", 3)
	   if len(parts) < 2 {
		   return
	   }
	   dest := parts[1]
	   logf("🔗 [%s] CONNECT para %s\n", remoteAddr, dest)
	   server, err := net.Dial("tcp", dest)
	   if err != nil {
		   client.Write([]byte("HTTP/1.1 502 Bad Gateway\r\n\r\n"))
		   logf("❌ [%s] Falha ao conectar em %s\n", remoteAddr, dest)
		   return
	   }
	   defer server.Close()
	   client.Write([]byte("HTTP/1.1 200 Connection Established\r\n\r\n"))
	   go io.Copy(server, client)
	   io.Copy(client, server)
	   logf("🔚 [%s] Tunnel CONNECT finalizado para %s\n", remoteAddr, dest)
	   return
   }

	// Requisição HTTP normal (GET, POST, ...)
	var host string
	var firstLine string
	if len(lines) > 0 {
		firstLine = lines[0]
	}
	for _, line := range lines {
		if strings.HasPrefix(strings.ToLower(line), "host:") {
			host = strings.TrimSpace(line[5:])
			break
		}
	}
   if host == "" {
	   logf("❌ [%s] Host não encontrado na requisição HTTP\n", remoteAddr)
	   return
   }
   logf("🌐 [%s] %s Host: %s\n", remoteAddr, firstLine, host)
   server, err := net.Dial("tcp", host+":80")
   if err != nil {
	   client.Write([]byte("HTTP/1.1 502 Bad Gateway\r\n\r\n"))
	   logf("❌ [%s] Falha ao conectar em %s:80\n", remoteAddr, host)
	   return
   }
   defer server.Close()
   // Repassar a requisição original
   server.Write(buf[:n])
   // Repassar resposta para o cliente
   io.Copy(client, server)
   logf("🔚 [%s] Requisição HTTP finalizada para %s\n", remoteAddr, host)
// ...existing code...
}

// logf imprime logs com horário centralizado
func logf(format string, a ...interface{}) {
	prefix := time.Now().Format("2006-01-02 15:04:05")
	fmt.Printf("%s %s", prefix, fmt.Sprintf(format, a...))
}