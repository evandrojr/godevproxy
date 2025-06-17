package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	var port string
	var mode string
	var socksVersion int
	var user, pass string
	flag.StringVar(&port, "port", "1080", "Porta do servidor")
	flag.StringVar(&mode, "mode", "socks", "Modo: socks ou http")
	flag.IntVar(&socksVersion, "socks-version", 5, "Versão do protocolo SOCKS (padrão 5)")
	flag.StringVar(&user, "user", "", "Usuário para autenticação (opcional)")
	flag.StringVar(&pass, "pass", "", "Senha para autenticação (opcional)")
	flag.Parse()

	// Se socksVersion não for informado ou for zero, usar 5 como padrão
	if socksVersion == 0 {
		socksVersion = 5
	}
	SocksVersion = byte(socksVersion)

   fmt.Println("🧪 GoDevProxy - Proxy Modular SOCKS5/HTTP")
   fmt.Println("==============================================")
	fmt.Println("Modo SOCKS5:  go run . --mode socks")
	fmt.Println("Modo HTTP:    go run . --mode http")
	fmt.Println("Para ativar autenticação, use as flags --user e --pass. Exemplo:")
	fmt.Println("  go run . --mode socks --user meuuser --pass minhasenha")
	fmt.Println("Se não informar, o proxy funcionará sem autenticação.")
	if user != "" && pass != "" {
		fmt.Printf("Usuário: %s | Senha: %s\n", user, pass)
		fmt.Printf("Exemplo SOCKS5: curl --socks5 %s:%s@localhost:%s https://g1.globo.com\n", user, pass, port)
		fmt.Printf("Exemplo HTTP:  curl -x http://%s:%s@localhost:%s https://g1.globo.com\n", user, pass, port)
	} else {
		fmt.Println("Sem autenticação (livre)")
		fmt.Printf("Exemplo SOCKS5: curl --socks5 localhost:%s https://g1.globo.com\n", port)
		fmt.Printf("Exemplo HTTP:  curl -x http://localhost:%s https://g1.globo.com\n", port)
	}
	fmt.Printf("Versão do SOCKS: %d\n", SocksVersion)
	fmt.Printf("Para definir a versão do SOCKS use: --socks-version <número> (ex: --socks-version 5)\n")

	switch mode {
	case "socks":
		if SocksVersion == 4 {
			server := NewSOCKS4Server(port)
			if err := server.Start(); err != nil {
				fmt.Fprintf(os.Stderr, "Erro fatal: %v\n", err)
				os.Exit(1)
			}
		} else {
			server := NewSOCKS5Server(port)
			SetAuthCredentials(user, pass)
			if err := server.Start(); err != nil {
				fmt.Fprintf(os.Stderr, "Erro fatal: %v\n", err)
				os.Exit(1)
			}
		}
	case "http":
		SetAuthCredentials(user, pass)
		if err := StartHTTPProxy(port); err != nil {
			fmt.Fprintf(os.Stderr, "Erro fatal: %v\n", err)
			os.Exit(1)
		}
	default:
		fmt.Fprintf(os.Stderr, "Modo inválido: %s. Use 'socks' ou 'http'.\n", mode)
		os.Exit(1)
	}
}
