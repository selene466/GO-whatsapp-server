// Package server
package server

import (
	"GO-whatsapp-server/src/whatsapp"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/skip2/go-qrcode"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
	"google.golang.org/protobuf/proto"
)

type Server struct {
	whatsapp *whatsapp.Whatsapp
}

type LoginResponse struct {
	Status      string `json:"status"`
	LoginStatus string `json:"login_status"`
}

type QRB64Response struct {
	Status string `json:"status"`
	QR     string `json:"qr_base64"`
}

type RestartResponse struct {
	Status string `json:"status"`
}

type SendResponse struct {
	Status string `json:"status"`
}

var serverPort string

func NewServer(wa *whatsapp.Whatsapp) *Server {
	return &Server{whatsapp: wa}
}

func (s *Server) Start() {
	http.HandleFunc("/api/status", func(w http.ResponseWriter, r *http.Request) {
		login := s.whatsapp.GetStatusLogin()
		loginRes := "logout"

		if login {
			loginRes = "login"
		}

		response := LoginResponse{
			Status:      "OK",
			LoginStatus: loginRes,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		json.NewEncoder(w).Encode(response)
	})

	http.HandleFunc("/api/get-qr", func(w http.ResponseWriter, r *http.Request) {
		qrRes := s.whatsapp.GetQRBase64()

		response := QRB64Response{
			Status: "OK",
			QR:     qrRes,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		json.NewEncoder(w).Encode(response)
	})

	http.HandleFunc("/api/get-qr-image-file", func(w http.ResponseWriter, r *http.Request) {
		qrRes := s.whatsapp.GetQRBase64()

		png, err := qrcode.Encode(qrRes, qrcode.Medium, 256)
		if err != nil {
			http.Error(w, "Failed to generate QR", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "image/png")
		w.WriteHeader(http.StatusOK)
		w.Write(png)
	})

	http.HandleFunc("/api/restart", func(w http.ResponseWriter, r *http.Request) {
		go func() {
			log.Println("Restarting WhatsApp client...")
			s.whatsapp.Stop()
			if err := s.whatsapp.Start(); err != nil {
				log.Printf("Error restarting WhatsApp client: %v", err)
			}
			log.Println("WhatsApp client restarted.")
		}()

		response := RestartResponse{
			Status: "OK",
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		json.NewEncoder(w).Encode(response)
	})

	http.HandleFunc("/api/send", func(w http.ResponseWriter, r *http.Request) {
		response := SendResponse{
			Status: "OK",
		}

		w.Header().Set("Content-Type", "application/json")

		phone := r.URL.Query().Get("phone_number")
		msg := r.URL.Query().Get("message")

		if phone == "" || msg == "" {
			response.Status = "Bad Request"
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(response)
			return
		}

		// JID example: 6281234567890@s.whatsapp.net
		jid, err := types.ParseJID(fmt.Sprintf("%s@%s", phone, types.DefaultUserServer))
		if err != nil {
			response.Status = "Invalid"
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(response)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		_, err = s.whatsapp.Client.SendMessage(ctx, jid, &waE2E.Message{
			Conversation: proto.String(msg),
		})
		if err != nil {
			response.Status = "Failed"
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(response)
			return
		}

		w.WriteHeader(http.StatusOK)

		json.NewEncoder(w).Encode(response)
	})

	if serverPort == "" {
		fmt.Println("Port not set in environment, using default 40040")
		serverPort = "40040"
	}

	log.Println("Starting server on http://0.0.0.0:" + serverPort)
	if err := http.ListenAndServe("0.0.0.0:"+serverPort, nil); err != nil {
		log.Fatalf("HTTP server error: %v", err)
	}
}
