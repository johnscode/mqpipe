package main

import (
	"github.com/rs/zerolog"
	"gorm.io/gorm"
)

type Repository struct {
	db     *gorm.DB
	logger *zerolog.Logger
}

func NewRepository(db *gorm.DB, logger *zerolog.Logger) *Repository {
	return &Repository{db: db, logger: logger}
}

func (r *Repository) Close() {
	sqlDb, err := r.db.DB()
	if err != nil {
		r.logger.Error().Err(err).Msg("failed to close database")
		return
	}
	_ = sqlDb.Close()
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

func (r *Repository) CreateMessage(message *IoTRawDeviceMessage) error {
	return r.db.Create(message).Error
}

func (r *Repository) GetMessagesByDeviceID(deviceID uint, limit int) ([]IoTRawDeviceMessage, error) {
	var messages []IoTRawDeviceMessage
	err := r.db.Where("device_id = ?", deviceID).Order("timestamp desc").Limit(limit).Find(&messages).Error
	return messages, err
}

func (r *Repository) DeleteMessagesByDeviceID(deviceID uint) error {
	return r.db.Where("device_id = ?", deviceID).Delete(&IoTRawDeviceMessage{}).Error
}

// Event Data-related functions

func (r *Repository) CreateDataEvent(message *IoTDeviceDataEvent) error {
	return r.db.Create(message).Error
}
