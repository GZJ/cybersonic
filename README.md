Cybersonic is a service for playing sound effects.

## quickstart
```shell
# start server
go run cybersonicd.go
go run cybersonicd.go --address [ip:port]

# list all sfx
go run cybersonic.go

# specify sfx
go run cybersonic.go beep
go run cybersonic.go beep --address [ip:port]
```
