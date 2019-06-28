# Blaulicht

Receive Prometheus alerts to let a hardware blue light attract attention from system operators.

## Test
To simulate a serial device run as root:
```bash
socat PTY,link=/dev/virtual-tty,mode=777,raw,echo=0 exec:'/bin/cat'
```

Send an alert:
```bash
curl -H "Authorization: Bearer $(cat token.key)" localhost:8080/api/v1/alert -d '{"receiver":"curl","alerts":[{"status":"create","labels":{"severity":"CRITICAL"}}]}'
```
