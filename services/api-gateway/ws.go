package main

import (
	"encoding/json"
	"log"
	"net/http"
	"ride-sharing/shared/contracts"
	"ride-sharing/shared/util"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func handleDriversWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		log.Printf("Websocket upgrade failed %v", err)
		return
	}

	defer conn.Close()

	userID := r.URL.Query().Get("userID")

	if userID == "" {
		log.Println("No userId found")
		return
	}

	packageSlug := r.URL.Query().Get("packageSlug")

	if packageSlug == "" {
		log.Println("No userId found")
		return
	}

	type Driver struct {
		Id             string `json:"id"`
		Name           string `json:"name"`
		ProfilePicture string `json:"profilePicture"`
		CarPlate       string `json:"carPlate"`
		PackageSlug    string `json:"packageSlug"`
	}

	driver := Driver{
		Id:             userID,
		Name:           "QD",
		ProfilePicture: util.GetRandomAvatar(1),
		CarPlate:       "AVCCC122",
		PackageSlug:    packageSlug,
	}

	driverData, err := json.Marshal(driver)
	if err != nil {
		log.Printf("Failed to marshal driver data: %v", err)
		return
	}

	msg := contracts.WSDriverMessage{
		Type: "driver.cmd.register",
		Data: driverData,
	}

	if err := conn.WriteJSON(msg); err != nil {
		log.Printf("WriteJSON failed %v", err)
		return

	}

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Error reading message %v", err)
			return
		}

		log.Printf("Received reading message %s", message)

	}

}

func handleRidersWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		log.Printf("Websocket upgrade failed %v", err)
		return
	}

	defer conn.Close()
	userID := r.URL.Query().Get("userID")

	if userID == "" {
		log.Println("No userId found")
		return
	} 

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Error reading message %v", err)
			return
		}

		log.Printf("Received reading message %s", message)

	}

}
