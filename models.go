package main

import (
	"encoding/json"
	"fmt"
	"time"
)

type BaseModel struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type DeviceModel struct {
	BaseModel
	DeviceID   string          `gorm:"uniqueIndex" json:"device_id"`
	DeviceName string          `json:"name"`
	DeviceType string          `json:"device_type"`
	Properties json.RawMessage `gorm:"type:json" json:"properties"`
}

func (d DeviceModel) DeviceId() string {
	return d.DeviceID
}

func (d DeviceModel) Name() string {
	return d.DeviceName
}

func (d DeviceModel) Type() string {
	return d.DeviceType
}

type TempRHDevice struct {
	DeviceModel
	Temp float32 `json:"temp,omitempty"`
	RH   float32 `json:"rh,omitempty"`
}

type IoTRawDeviceMessage struct {
	BaseModel
	Time       time.Time       `json:"time"`
	DeviceID   string          `json:"device_id"`
	DeviceType string          `json:"device_type"`
	DeviceData json.RawMessage `json:"device_data"`
}

type IoTDeviceDataEvent struct {
	BaseModel
	Time       time.Time `json:"time" gorm:"index"`
	DeviceID   string    `json:"device_id"`
	DeviceType string    `json:"device_type"`
	DeviceData string    `json:"device_data" gorm:"type:jsonb"`
}

func (m IoTRawDeviceMessage) Format(f fmt.State, _ rune) {
	_, _ = fmt.Fprintf(f, "IotDeviceMessage{\n\tid: %d,\n\tCreatedAt: %s,\n\tUpdatedAt: %s,\n\tTime: %s,\n\tDeviceId: %s,\n\tDeviceType: %s,\n\tProperties: %s,\n}",
		m.ID, m.CreatedAt, m.UpdatedAt, m.Time, m.DeviceID, m.DeviceType, m.DeviceData)
}
