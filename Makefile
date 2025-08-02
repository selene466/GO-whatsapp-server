ifeq (,$(wildcard .env))
 $(error .env file is missing)
endif

include .env
export

BINARY_PATH := $(HOME)/.local/bin/$$GO_WHATSAPP_SERVER_NAME
SERVICE_DIR := $(HOME)/.config/systemd/user
SERVICE_FILE := $(SERVICE_DIR)/$$GO_WHATSAPP_SERVER_NAME.service


build:
	mkdir -p $(dir $(BINARY_PATH))
	go build -ldflags "-X GO-whatsapp-server/src/server.serverPort=$$GO_WHATSAPP_SERVER_PORT" -o $(BINARY_PATH) ./main.go

install-service: $(SERVICE_FILE)

$(SERVICE_FILE):
	mkdir -p $(SERVICE_DIR)
	@echo "[Unit]"                          >  $(SERVICE_FILE)
	@echo "Description=$$GO_WHATSAPP_SERVER_NAME"        >> $(SERVICE_FILE)
	@echo "After=network.target"           >> $(SERVICE_FILE)
	@echo ""                               >> $(SERVICE_FILE)
	@echo "[Service]"                      >> $(SERVICE_FILE)
	@echo "ExecStart=$(BINARY_PATH)"       >> $(SERVICE_FILE)
	@echo "Restart=always"                 >> $(SERVICE_FILE)
	@echo "Environment=ENV=production"     >> $(SERVICE_FILE)
	@echo ""                               >> $(SERVICE_FILE)
	@echo "[Install]"                      >> $(SERVICE_FILE)
	@echo "WantedBy=default.target"        >> $(SERVICE_FILE)

reload:
	systemctl --user daemon-reexec || systemctl --user daemon-reload

enable:
	systemctl --user enable $$GO_WHATSAPP_SERVER_NAME.service

disable:
	systemctl --user disable $$GO_WHATSAPP_SERVER_NAME.service

start:
	systemctl --user start $$GO_WHATSAPP_SERVER_NAME.service

stop:
	systemctl --user stop $$GO_WHATSAPP_SERVER_NAME.service

restart:
	systemctl --user restart $$GO_WHATSAPP_SERVER_NAME.service

status:
	systemctl --user status $$GO_WHATSAPP_SERVER_NAME.service

logs:
	journalctl --user -u $$GO_WHATSAPP_SERVER_NAME.service -f

clean:
	rm -f $(BINARY_PATH)
	rm -f $(SERVICE_FILE)

.PHONY: build install-service reload enable disable start stop restart status logs clean
