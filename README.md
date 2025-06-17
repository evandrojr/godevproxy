
# GoDevProxy - Proxy Modular SOCKS5/HTTP em Go

Este projeto implementa o GoDevProxy, um servidor proxy que pode operar tanto no modo SOCKS5 quanto HTTP, ambos com autenticação de usuário e senha.

## Funcionalidades
- Proxy SOCKS5 com autenticação USERNAME/PASSWORD
- Proxy HTTP com autenticação básica
- Logs detalhados de conexões e transferências

## Como usar

### Pré-requisitos
- Go instalado (versão 1.x ou superior)

### Executando o servidor

Execute diretamente (GoDevProxy SOCKS5 é padrão):

```bash
go run . --mode socks   # Para SOCKS5 (padrão)
go run . --mode http    # Para Proxy HTTP
go run . --mode socks --port 1080
```

### Exemplo de uso com curl (GoDevProxy)

**SOCKS5:**
```bash
curl --socks5 $USUARIO:$SENHA@localhost:1080 https://g1.globo.com
```

**HTTP Proxy:**
```bash
curl -x http://$USUARIO:$SENHA@localhost:1080 https://g1.globo.com
```

Se não informar usuário e senha via CLI, use os valores padrão:

Usuário: `admin`  
Senha: `123`

Para definir usuário e senha personalizados, rode o servidor assim:

```bash
go run . --mode socks --user meuuser --pass minhasenha
```
E use no curl:
```bash
curl --socks5 meuuser:minhasenha@localhost:1080 https://g1.globo.com
```

Por exemplo, para rodar na porta 8080:

```bash
./socks5-server 8080
```

O GoDevProxy exibirá logs no console indicando que está pronto para aceitar conexões.

### Testing the server

Você pode testar o GoDevProxy configurando um cliente SOCKS5 para usar `localhost` e a porta em que o servidor está ouvindo.

Por exemplo, usando `curl`:

```bash
curl --socks5 localhost:1080 https://google.com
```

Substitua `1080` pela porta que você especificou, se diferente.

## Code Structure

-   [`socks5_server.go`](socks5_server.go ): Contains all the SOCKS5 server logic.
    -   [`main()`](socks5_server.go ): Entry point, parses command-line arguments and starts the server.
    -   `SOCKS5Server`: Struct representing the server.
    -   [`NewSOCKS5Server()`](socks5_server.go ): Creates a new instance of `SOCKS5Server`.
    -   `Start()`: Starts the TCP listener and accepts connections.
    -   `handleConnection()`: Manages each client connection, including authentication and request processing.
    -   `handleAuth()`: Handles the SOCKS5 authentication negotiation phase.
    -   `handleRequest()`: Processes the client's SOCKS5 connection request.
    -   `sendErrorResponse()`: Sends a SOCKS5 error response to the client.
    -   `relay()`: Forwards data bidirectionally between the client and the destination.

## How SOCKS5 Works (Simplified)

1.  **Authentication Negotiation**:
    *   The client connects to the SOCKS5 server and sends a message indicating the SOCKS version and supported authentication methods.
    *   The server chooses one of the methods (in this case, "no authentication") and sends a response to the client.

2.  **Connection Request**:
    *   The client sends a request to the server specifying the command (e.g., CONNECT), the destination address type (IPv4, domain, IPv6), and the destination address/port.
    *   The server attempts to connect to the requested destination.

3.  **Server Response**:
    *   The server sends a response to the client indicating whether the connection was successful or if an error occurred. If successful, the response includes the address and port that the server is using to connect to the destination (usually not the same as the client requested, but rather the proxy's own address).

4.  **Data Forwarding**:
    *   If the connection is successfully established, the SOCKS5 server begins to forward data bidirectionally between the client and the destination server.

O GoDevProxy implementa as fases acima para o método "no authentication" e o comando "CONNECT".
