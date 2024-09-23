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

type IoTDeviceMessage struct {
	BaseModel
	Time     time.Time `json:"time"`
	DeviceID string    `json:"device_id"`
	//Device   DeviceModel `gorm:"polymorphic:Device;" json:"device"`
	Device DeviceModel `gorm:"embedded" json:"device"`
}

func (m *IoTDeviceMessage) UnmarshalJSON(data []byte) error {
	var raw struct {
		Time   time.Time       `json:"time"`
		Device json.RawMessage `json:"device"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	m.Time = raw.Time
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
		properties, _ := json.Marshal(struct {
			Temp float32 `json:"temp,omitempty"`
			RH   float32 `json:"rh,omitempty"`
		}{
			Temp: tempRHDevice.Temp,
			RH:   tempRHDevice.RH,
		})
		m.Device.Properties = properties
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
		var tempRH TempRHDevice
		tempRH.DeviceModel = m.Device
		_ = json.Unmarshal(m.Device.Properties, &tempRH)
		data, _ := json.Marshal(tempRH)
		return data
	default:
		data, _ := json.Marshal(m.Device)
		return data
	}
}

func (m IoTDeviceMessage) Format(f fmt.State, _ rune) {
	props, _ := json.Marshal(m.Device.Properties)
	_, _ = fmt.Fprintf(f, "IotDeviceMessage{\n\tid: %d,\n\tCreatedAt: %s,\n\tUpdatedAt: %s,\n\tTime: %s,\n\tDeviceId: %s,\n\tDeviceName: %s,\n\tDeviceType: %s,\n\tProperties: %s,\n}",
		m.ID, m.CreatedAt, m.UpdatedAt, m.Time, m.DeviceID, m.Device.DeviceName, m.Device.DeviceType, props)
}
