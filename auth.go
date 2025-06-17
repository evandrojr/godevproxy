package main

import (
	"net"
	"fmt"
)

var authUser string
var authPass string

func SetAuthCredentials(user, pass string) {
   authUser = user
   authPass = pass
}

func CheckUserPass(username, password string) bool {
   if authUser == "" && authPass == "" {
	   // Sem autenticação
	   return true
   }
   return username == authUser && password == authPass
}

// SOCKS5 authentication negotiation
func HandleSocks5Auth(conn net.Conn, clientAddr string) bool {
	buf := make([]byte, 256)
	n, err := conn.Read(buf)
	if err != nil {
	   fmt.Printf("❌ [%s] Erro ao ler auth request: %v\n", clientAddr, err)
		return false
	}
   if n < 3 || buf[0] != SocksVersion {
	   fmt.Printf("❌ [%s] Versão SOCKS inválida: %02x\n", clientAddr, buf[0])
	   return false
   }
	nmethods := int(buf[1])
	if n < 2+nmethods {
	   fmt.Printf("❌ [%s] Requisição de auth malformada\n", clientAddr)
		return false
	}
   supportUserPass := false
   for i := 0; i < nmethods; i++ {
	   if buf[2+i] == USERPASS_AUTH {
		   supportUserPass = true
		   break
	   }
   }
   if !supportUserPass {
	   conn.Write([]byte{SocksVersion, 0xFF})
	   fmt.Printf("❌ [%s] Cliente não suporta USERNAME/PASSWORD\n", clientAddr)
	   return false
   }
   conn.Write([]byte{SocksVersion, USERPASS_AUTH})
	n, err = conn.Read(buf)
	if err != nil || n < 5 {
	   fmt.Printf("❌ [%s] Erro ao ler user/pass: %v\n", clientAddr, err)
		return false
	}
	if buf[0] != USERPASS_VERSION {
	   fmt.Printf("❌ [%s] Versão USERPASS inválida: %02x\n", clientAddr, buf[0])
		return false
	}
	ulen := int(buf[1])
	if n < 2+ulen+1 {
	   fmt.Printf("❌ [%s] USERNAME incompleto\n", clientAddr)
		return false
	}
	username := string(buf[2 : 2+ulen])
	plen := int(buf[2+ulen])
	if n < 2+ulen+1+plen {
	   fmt.Printf("❌ [%s] PASSWORD incompleto\n", clientAddr)
		return false
	}
	password := string(buf[3+ulen : 3+ulen+plen])
	if CheckUserPass(username, password) {
		conn.Write([]byte{USERPASS_VERSION, AUTH_SUCCESS})
		fmt.Printf("✅ [%s] Autenticado como '%s'\n", clientAddr, username)
		return true
	} else {
		conn.Write([]byte{USERPASS_VERSION, AUTH_FAILURE})
	   fmt.Printf("❌ [%s] Falha na autenticação para '%s'\n", clientAddr, username)
		return false
	}
}
