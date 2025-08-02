// Package whatsapp
package whatsapp

import (
	"GO-whatsapp-server/src/dbsqlite"
	"context"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"
	"google.golang.org/protobuf/proto"

	"github.com/adrg/xdg"
	"github.com/mdp/qrterminal/v3"
)

type Whatsapp struct {
	ctx      context.Context
	Client   *whatsmeow.Client
	running  bool
	loggedin bool
	qrbase64 string
	mu       sync.Mutex
}

func newLogger(name string) waLog.Logger {
	return waLog.Stdout(name, "INFO", true)
}

var Log = newLogger("GO-whatsapp-server")

func GetLogger() waLog.Logger {
	return Log
}

func (w *Whatsapp) eventHandler(evt interface{}) {
	switch v := evt.(type) {
	case *events.Message:
		sender := v.Info.Sender
		senderPhone := strings.Split(sender.String(), "@")[0]
		Log.Infof("Received a message:\n%s\n%s", senderPhone, v.Message.GetConversation())

		if v.Message.ImageMessage != nil && v.Message.ImageMessage.Caption != nil {

			// 1. Capture image message
			fileExt := ".jpg"
			fileString := xdg.CacheHome + "GO-whatsapp-server-" + v.Info.ID
			filePath := fileString + fileExt
			imageData, err := w.Client.Download(w.ctx, v.Message.GetImageMessage())
			if err != nil {
				Log.Errorf("Failed to save image: %v", err)
				return
			}

			// 2. Save image
			err = os.WriteFile(filePath, imageData, 0644)
			if err != nil {
				Log.Errorf("Failed to save image:", err)
				return
			}

			// 3. Do something with the image
			// Received image with caption here
			Log.Infof("Received image with caption: %s", *v.Message.ImageMessage.Caption)
			Log.Infof("Image saved to: %s", filePath)
			time.Sleep(time.Second * 3)

			// 4. Delete processed image
			matches, err := filepath.Glob(fileString + "*")
			if err != nil {
				return
			}

			for _, match := range matches {
				err = os.RemoveAll(match)
				if err != nil {
					return
				}
			}

		} else {
			// Received message here
			Log.Infof(v.Message.GetConversation())
		}
	}
}

func NewWhatsapp(ctx context.Context) *Whatsapp {
	return &Whatsapp{
		ctx:      ctx,
		Client:   nil,
		running:  false,
		loggedin: false,
		qrbase64: "",
	}
}

func (w *Whatsapp) Start() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	pattern := filepath.Join(xdg.CacheHome, "GO-whatsapp-server-*")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		panic(err)
	}

	for _, path := range matches {
		Log.Infof("Removing cache: " + path)
		if err := os.RemoveAll(path); err != nil {
			Log.Errorf("Failed to delete: " + path)
		}
	}

	if w.Client != nil && w.Client.IsConnected() {
		Log.Infof("Whatsapp client already running, disconnecting existing client.")
		w.Client.Disconnect()
		w.Client = nil
	}

	w.loggedin = false
	w.qrbase64 = ""

	dbLog := waLog.Stdout("Database", "DEBUG", true)
	ctx := context.Background()
	dbFilePath, err := dbsqlite.FileDB()
	if err != nil {
		return err
	}
	container, err := sqlstore.New(ctx, "sqlite3", "file:"+dbFilePath+"?_foreign_keys=on", dbLog)
	if err != nil {
		return err
	}

	store.DeviceProps.Os = proto.String("GO-whatsapp-server-selene466")
	deviceStore, err := container.GetFirstDevice(ctx)
	if err != nil {
		return err
	}
	clientLog := waLog.Stdout("Client", "DEBUG", true)
	w.Client = whatsmeow.NewClient(deviceStore, clientLog)
	w.Client.AddEventHandler(w.eventHandler)

	if w.Client.Store.ID == nil {
		// No Login ID stored, new login
		qrChan, _ := w.Client.GetQRChannel(context.Background())
		err = w.Client.Connect()
		if err != nil {
			return err
		}
		for evt := range qrChan {
			if evt.Event == "code" {
				w.loggedin = false
				w.qrbase64 = evt.Code
				Log.Infof("QR code: %s", evt.Code)
				qrterminal.GenerateHalfBlock(string(evt.Code), qrterminal.L, os.Stdout)
			} else if evt.Event == "success" {
				w.loggedin = true
				w.qrbase64 = ""
				Log.Infof("Login event: %s", evt.Event)
				break
			} else {
				Log.Infof("Login event: %s", evt.Event)
			}
		}
	} else {
		// Already logged in, just connect
		err = w.Client.Connect()
		if err != nil {
			return err
		}
		w.loggedin = true
	}

	w.running = true
	Log.Infof("Whatsapp client started.")
	return nil
}

func (w *Whatsapp) Stop() {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.Client != nil && w.Client.IsConnected() {
		Log.Infof("Whatsapp client stopping.")
		w.Client.Disconnect()
		w.Client = nil
	}
	w.running = false
	w.loggedin = false
	w.qrbase64 = ""
	Log.Infof("Whatsapp client stopped.")
}

func (w *Whatsapp) GetQRBase64() string {
	return w.qrbase64
}

func (w *Whatsapp) GetStatusLogin() bool {
	return w.loggedin
}
