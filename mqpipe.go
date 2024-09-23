package main

import (
	"context"
	"encoding/json"
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/rs/zerolog"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
)

const (
	broker   = "tcp://localhost:1883"
	clientID = "go-mqtt-subscriber"
	topic    = "iot-messages"
)

// object structure for dealing with messages from MQTT broker
//
//type MQTTDeviceMessage struct {
//	Time   time.Time  `json:"time"`
//	Device MQTTDevice `json:"device"`
//}
//
//func (m *MQTTDeviceMessage) UnmarshalJSON(data []byte) error {
//
//	//   to determine device type then unmarshal into proper type
//	var raw rawMqttTempRHDeviceMessage
//	if err := json.Unmarshal(data, &raw); err != nil {
//		return err
//	}
//	m.Time = raw.Time
//	m.Device = &raw.Device
//	return nil
//}
//
//type rawMqttTempRHDeviceMessage struct {
//	Time   time.Time        `json:"time"`
//	Device MQTTTempRHDevice `json:"device"`
//}
//
//type MQTTDevice interface {
//	ID() string
//	Name() string
//}
//
//type MQTTTempRHDevice struct {
//	Id         string  `json:"id"`
//	DeviceName string  `json:"name,omitempty"`
//	Temp       float32 `json:"temp,omitempty"`
//	Rh         float32 `json:"rh,omitempty"`
//}
//
//func (t MQTTTempRHDevice) ID() string {
//	return t.Id
//}
//
//func (t MQTTTempRHDevice) Name() string {
//	return t.DeviceName
//}

var mqttMsgChan = make(chan mqtt.Message)

var messagePubHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	mqttMsgChan <- msg
}

func processMsg(ctx context.Context, logger *zerolog.Logger, input <-chan mqtt.Message) chan IoTDeviceMessage {
	out := make(chan IoTDeviceMessage)
	go func() {
		defer close(out)
		for {
			select {
			case msg, ok := <-input:
				if !ok {
					return
				}
				fmt.Printf("Received message: %s from topic: %s\n", msg.Payload(), msg.Topic())
				var iotMsg IoTDeviceMessage
				err := json.Unmarshal(msg.Payload(), &iotMsg)
				if err != nil {
					logger.Error().Err(err).Msg("Error unmarshalling IoTDeviceMessage")
				} else {
					out <- iotMsg
				}
			case <-ctx.Done():
				return
			}
		}
	}()
	return out
}

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	fmt.Println("Connected to MQTT Broker")
}

var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	fmt.Printf("Connection lost: %v", err)
}

func main() {
	appCtx := context.Background()
	logger := setupLogger(appCtx, "")

	_ = setupPostgres(logger)

	opts := mqtt.NewClientOptions()
	opts.AddBroker(broker)
	opts.SetClientID(clientID)
	opts.SetDefaultPublishHandler(messagePubHandler)
	opts.OnConnect = connectHandler
	opts.OnConnectionLost = connectLostHandler

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	ctx, cancel := context.WithCancel(appCtx)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		finalChan := processMsg(ctx, logger, mqttMsgChan)
		for iotMsg := range finalChan {
			// now we have the mqtt message as parsed from json
			logger.Info().Msg(fmt.Sprintf("Received iot msg: %+v", iotMsg))
			// do something like save to db
		}
	}()

	// Subscribe to the topic
	token := client.Subscribe(topic, 1, nil)
	token.Wait()
	fmt.Printf("Subscribed to topic: %s\n", topic)

	// Wait for interrupt signal to gracefully shutdown the subscriber
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	// Cancel the context to signal the goroutine to stop
	cancel()

	// Unsubscribe and disconnect
	fmt.Println("Unsubscribing and disconnecting...")
	client.Unsubscribe(topic)
	client.Disconnect(250)

	// Wait for the goroutine to finish
	wg.Wait()
	fmt.Println("Goroutine terminated, exiting...")
}

func setupLogger(ctx context.Context, logFilePath string) *zerolog.Logger {
	var outWriter = os.Stdout
	if logFilePath != "" && logFilePath != "stdout" {
		file, err := os.OpenFile(logFilePath,
			os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0664)
		if err != nil {
			log.Fatalln(err)
		}
		outWriter = file
	}
	cout := zerolog.ConsoleWriter{Out: outWriter, TimeFormat: time.RFC822}
	cout.FormatLevel = func(i interface{}) string {
		return strings.ToUpper(fmt.Sprintf("| %-6s|", i))
	}
	// uncomment to remove timestamp from logs
	//out.FormatTimestamp = func(i interface{}) string {
	//	return ""
	//}
	baseLogger := zerolog.New(cout).With().Timestamp().Logger()
	logCtx := baseLogger.WithContext(ctx)
	l := zerolog.Ctx(logCtx)
	return l
}
