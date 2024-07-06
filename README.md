# Shadowtracker

This tool interprets any HTTP or HTTPS traffic of a specified port and forwards it to another port, after getting the response it will return the response to the client. This tools monitors the http and https traffic forwards them to specified ports and also logs the traffic on a log file for future inspection.

> As of now this tool is only available for systems with Go programming language installed

## Installation

### Install Go

Follow the [official documentation](https://go.dev/doc/install) to install Go on your system.

```bash
go install github.com/sifatulrabbi/shadowtracker@latest
```

### Run shadowtracker and monitor your traffic

Shadowtracker is not a HTTP server it's a TCP interceptor that acts as the middleman between the client and your HTTP server. Assuming that you have your HTTP server running on the same machine.

- Specify the port number where your http server is currently running with `-forward`
- Specify the outbound port with `-target`. Specify the logs dir with `-logsdir`
- If you do not specify this then the default destination will be `/var/logs/shadowtracker`
- If you use anything outside your `$HOME` directory in linux then you need to use `sudo`

```bash
shadowtracker -target 80 -forward 8000 -logsdir /home/root/logs
```

**With sudo**

```bash
sudo shadowtracker -target 80 -forward 8000 -logsdir /var/logs/my-server
```
