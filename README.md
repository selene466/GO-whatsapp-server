# whatsapp-server

Ready to use WhatsApp server from WhatsApp Web API.

Thanks to:

- [whatsmeow](https://github.com/tulir/whatsmeow)

This is rewrite version of [BUN-whatsapp-server](https://github.com/selene466/BUN-whatsapp-server).

## How to install

Install build tools.

Ubuntu:

```sh
sudo apt install build-essential
```

Arch:

```sh
sudo pacman -S base-devel
```

Setup [GO](https://go.dev/doc/install) & `$GOPATH` to your environment:

```sh
export GOPATH=$HOME/go
```

Clone this repo.

Install dependencies:

```sh
go mod tidy
```

Copy `.env.example` to `.env` and set the values.

To install:

```sh
make build
```

Default port is `40040`.

To uninstall:

```sh
make clean
```

To install systemd service:

```sh
make install-service
```

To start systemd service:

```sh
make start
```

To stop systemd service:

```sh
make stop
```

To enable systemd service on boot:

```sh
make enable
```

To disable systemd service on boot:

```sh
make disable
```

To view systemd service status:

```sh
make status
```

To view journalctl logs:

```sh
make logs
```

## API Docs

Each login QR is generated with timeout of ~1 minute.  
WhatsApp Web API will close the connection after 3 login QR not scanned.  
If this happen, hit API restart manually.

Login status:

```sh
curl -X GET http://localhost:40040/api/status
```

Send message, message must be encoded in URL (URL encode):

```sh
curl -X GET \
http://localhost:40040/api/send?phone_number=6285100001234&message=Hello%20World
```

Get login QR base64:

```sh
curl -X GET http://localhost:40040/api/get-qr
```

Get login QR image PNG:

```sh
curl -X GET http://localhost:40040/api/get-qr-image-file
```

## Incoming Message Handler

Add your own handler in `src/whatsapp/whatsapp.go`.

```node
// Received message here
```
