package internal

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

func SetupMQTT(config Config, db *sql.DB) {
	// MQTT setup
	log.Println("Configuring MQTT client...")
	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("%s://%s:%d", config.MQTT.Protocol, config.MQTT.Host, config.MQTT.Port))
	opts.SetUsername(config.MQTT.User)
	opts.SetPassword(config.MQTT.Pass)
	opts.SetClientID("go-mqtt-subscriber")
	opts.SetAutoReconnect(true)
	opts.SetConnectRetryInterval(10 * time.Second)    // Retry mỗi 10 giây ban đầu
	opts.SetMaxReconnectInterval(30 * 24 * time.Hour) // Tăng lên tối đa 1 tháng giữa các lần retry

	var client mqtt.Client
	for {
		client = mqtt.NewClient(opts)
		log.Println("Connecting to MQTT broker...")
		if token := client.Connect(); token.Wait() && token.Error() != nil {
			log.Printf("MQTT connection failed: %v. Retrying in 5 seconds...", token.Error())
			time.Sleep(5 * time.Second)
			continue
		}
		log.Println("Connected to MQTT broker successfully.")
		break
	}
	// defer client.Disconnect(250) // REMOVED: Don't disconnect immediately

	// Subscribe to topic
	log.Printf("Subscribing to topic: %s", config.MQTT.Topic)
	for {
		token := client.Subscribe(config.MQTT.Topic, 0, func(client mqtt.Client, msg mqtt.Message) {
			log.Printf("DEBUG: Callback triggered for topic %s", msg.Topic())
			log.Printf("Received message on topic %s: %s", msg.Topic(), string(msg.Payload()))
			if strings.Contains(msg.Topic(), "telemetry") {
				ProcessTelemetryMessage(db, msg)
			} else if strings.Contains(msg.Topic(), "attributes") {
				ProcessAttributesMessage(db, msg)
			} else {
				log.Printf("DEBUG: Topic %s does not contain 'telemetry' or 'attributes'", msg.Topic())
			}
		})
		token.Wait()
		if token.Error() != nil {
			log.Printf("Subscription failed: %v. Retrying in 5 seconds...", token.Error())
			time.Sleep(5 * time.Second)
			continue
		}
		log.Println("Subscription successful. Waiting for messages...")
		break
	}
}
