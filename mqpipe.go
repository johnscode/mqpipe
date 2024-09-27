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

var mqttMsgChan = make(chan mqtt.Message)

var messagePubHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	mqttMsgChan <- msg
}

func processMsg(ctx context.Context, logger *zerolog.Logger, input <-chan mqtt.Message) chan IoTRawDeviceMessage {
	out := make(chan IoTRawDeviceMessage)
	go func() {
		defer close(out)
		for {
			select {
			case msg, ok := <-input:
				if !ok {
					return
				}
				fmt.Printf("Received message: %s from topic: %s\n", msg.Payload(), msg.Topic())
				var iotMsg IoTRawDeviceMessage
				err := json.Unmarshal(msg.Payload(), &iotMsg)
				if err != nil {
					logger.Error().Err(err).Msg("Error unmarshalling IoTRawDeviceMessage")
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

func persistIoTEvent(ctx context.Context, logger *zerolog.Logger, repo *Repository, input <-chan IoTRawDeviceMessage) chan IoTRawDeviceMessage {
	out := make(chan IoTRawDeviceMessage)
	go func() {
		defer close(out)
		for iotMsg := range input {
			logger.Info().Msg(fmt.Sprintf("Persist iot msg for device: %s", iotMsg.DeviceID))
			msg := IoTDeviceDataEvent{
				BaseModel:  iotMsg.BaseModel,
				Time:       iotMsg.Time,
				DeviceID:   iotMsg.DeviceID,
				DeviceType: iotMsg.DeviceType,
				DeviceData: string(iotMsg.DeviceData),
			}
			err := repo.CreateDataEvent(&msg)
			if err != nil {
				logger.Error().Err(err).Msg("Error creating IoTRawDeviceMessage")
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

	repo := setupPostgres(logger)

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
		finalChan := persistIoTEvent(ctx, logger, repo, processMsg(ctx, logger, mqttMsgChan))
		for iotMsg := range finalChan {
			// now we have the IoTRawDeviceMessage that has been persisted
			logger.Info().Msg(fmt.Sprintf("Received iot msg: %+v", iotMsg))
			// do something like check for alert conditions
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
	// finally close the db connection
	repo.Close()
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
