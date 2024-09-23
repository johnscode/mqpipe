package main

import (
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// Device-related functions

func (r *Repository) CreateTempRHDevice(device *TempRHDevice) error {
	return r.db.Create(&device).Error
}

func (r *Repository) GetTempRHDeviceByDeviceID(deviceID string) (*TempRHDevice, error) {
	var device TempRHDevice
	err := r.db.Where("device_id = ?", deviceID).First(&device).Error
	if err != nil {
		return nil, err
	}
	return &device, nil
}

func (r *Repository) UpdateTempRHDevice(device *TempRHDevice) error {
	return r.db.Save(device).Error
}

func (r *Repository) DeleteTempRHDevice(deviceID string) error {
	return r.db.Where("device_id = ?", deviceID).Delete(&TempRHDevice{}).Error
}

// Message-related functions

func (r *Repository) CreateMessage(message *IoTDeviceMessage) error {
	return r.db.Create(message).Error
}

func (r *Repository) GetMessagesByDeviceID(deviceID uint, limit int) ([]IoTDeviceMessage, error) {
	var messages []IoTDeviceMessage
	err := r.db.Where("device_id = ?", deviceID).Order("timestamp desc").Limit(limit).Find(&messages).Error
	return messages, err
}

func (r *Repository) DeleteMessagesByDeviceID(deviceID uint) error {
	return r.db.Where("device_id = ?", deviceID).Delete(&IoTDeviceMessage{}).Error
}
