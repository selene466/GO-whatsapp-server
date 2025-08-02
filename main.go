package main

import (
	"GO-whatsapp-server/src/dbsqlite"
	"GO-whatsapp-server/src/server"
	"GO-whatsapp-server/src/whatsapp"
	"context"
	"log"
)

func main() {
	// Init dbsqlite
	if err := dbsqlite.Init(); err != nil {
		panic(err)
	}

	ctx := context.Background()
	wa := whatsapp.NewWhatsapp(ctx)

	// Start the HTTP server
	srv := server.NewServer(wa)
	go srv.Start()

	// Start the WhatsApp client
	if err := wa.Start(); err != nil {
		log.Printf("Failed to start WhatsApp client initially: %v", err)
	}

	// Keep goroutine alive
	select {}
}
