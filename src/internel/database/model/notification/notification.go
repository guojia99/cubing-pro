package notification

import "time"

type SiteAnnouncementRead struct {
	UserID         uint      `gorm:"primaryKey;autoIncrement:false"`
	AnnouncementID string    `gorm:"primaryKey;type:varchar(36)"`
	ReadAt         time.Time `gorm:"not null"`
}
