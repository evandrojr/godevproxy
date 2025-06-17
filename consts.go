package main

var SocksVersion byte = 0x05

const (
   NO_AUTH           = 0x00
   USERPASS_AUTH     = 0x02
   USERPASS_VERSION  = 0x01
   AUTH_SUCCESS      = 0x00
   AUTH_FAILURE      = 0x01
   CONNECT           = 0x01
   IPV4              = 0x01
   DOMAIN            = 0x03
   SUCCESS           = 0x00
   FAILURE           = 0x01
)
