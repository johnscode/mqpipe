package main

import (
	"encoding/json"
	"time"
)

type BaseModel struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Device interface {
	ID() string
	Name() string
	Type() string
}

type DeviceModel struct {
	BaseModel
	DeviceID   string `gorm:"uniqueIndex" json:"device_id"`
	DeviceName string `json:"name"`
	DeviceType string `json:"device_type"`
}

func (d DeviceModel) ID() string {
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

type rawTempRHDeviceMessage struct {
	Time   time.Time    `json:"time"`
	Device TempRHDevice `json:"device"`
}

type IoTDeviceMessage struct {
	BaseModel
	Time     time.Time   `json:"time"`
	DeviceID string      `json:"device_id"`
	Device   DeviceModel `gorm:"polymorphic:Device;" json:"device"`
}

func (m *IoTDeviceMessage) UnmarshalJSON(data []byte) error {
	var raw struct {
		Time   time.Time       `json:"time"`
		Device json.RawMessage `json:"device"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	var baseDevice DeviceModel
	if err := json.Unmarshal(raw.Device, &baseDevice); err != nil {
		return err
	}
	switch baseDevice.DeviceType {
	case "TempRH":
		var tempRHDevice TempRHDevice
		if err := json.Unmarshal(raw.Device, &tempRHDevice); err != nil {
			return err
		}
		m.Device = tempRHDevice.DeviceModel
	default:
		m.Device = baseDevice
	}

	m.DeviceID = m.Device.DeviceID
	return nil
}

func (m IoTDeviceMessage) MarshalJSON() ([]byte, error) {
	type Alias IoTDeviceMessage
	return json.Marshal(&struct {
		Alias
		Device json.RawMessage `json:"device"`
	}{
		Alias:  Alias(m),
		Device: m.marshalDevice(),
	})
}

func (m IoTDeviceMessage) marshalDevice() json.RawMessage {
	switch m.Device.DeviceType {
	case "TempRH":
		tempRH := TempRHDevice{DeviceModel: m.Device}
		// set Temp and RH here if needed
		data, _ := json.Marshal(tempRH)
		return data
	default:
		data, _ := json.Marshal(m.Device)
		return data
	}
}
