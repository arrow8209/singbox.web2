package storage

import (
	"time"
)

type Setting struct {
	Key       string `gorm:"primaryKey"`
	Value     string
	UpdatedAt time.Time
}

type Inbound struct {
	ID        uint   `gorm:"primaryKey"`
	Name      string `gorm:"not null"`
	Type      string `gorm:"not null"`
	Config    string `gorm:"not null"` // JSON
	Enabled   bool   `gorm:"default:true"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Subscription struct {
	ID             uint   `gorm:"primaryKey"`
	Name           string `gorm:"not null"`
	URL            string `gorm:"not null"`
	Type           string // auto/singbox/clash/v2ray/base64
	UpdateInterval int    // hours
	LastUpdate     *time.Time
	Enabled        bool `gorm:"default:true"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type Outbound struct {
	ID             uint   `gorm:"primaryKey"`
	SubscriptionID *uint  // NULL for manual
	Name           string `gorm:"not null"`
	Type           string `gorm:"not null"`
	Server         string
	Port           int
	Config         string `gorm:"not null"` // JSON
	Latency        *int   // ms
	Enabled        bool   `gorm:"default:true"`
	CreatedAt      time.Time
	UpdatedAt      time.Time

	Subscription *Subscription `gorm:"foreignKey:SubscriptionID"`
}

type Ruleset struct {
	ID             uint   `gorm:"primaryKey"`
	Name           string `gorm:"not null"`
	Type           string `gorm:"not null"` // remote/local
	Format         string `gorm:"not null"` // source/binary
	URL            string
	Path           string // local file path
	UpdateInterval int    // hours
	LastUpdate     *time.Time
	Enabled        bool `gorm:"default:true"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type Rule struct {
	ID          uint   `gorm:"primaryKey"`
	Priority    int    `gorm:"not null"`
	Type        string `gorm:"not null"` // domain/ip/geoip/geosite/ruleset...
	Value       string `gorm:"not null"`
	OutboundTag string `gorm:"not null"`
	Enabled     bool   `gorm:"default:true"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type OperationLog struct {
	ID        uint   `gorm:"primaryKey"`
	Action    string `gorm:"not null"`
	Detail    string
	CreatedAt time.Time
}

type TrafficStat struct {
	ID         uint   `gorm:"primaryKey"`
	TargetType string `gorm:"not null"` // outbound/inbound/rule
	TargetName string `gorm:"not null"`
	Upload     int64  `gorm:"default:0"`
	Download   int64  `gorm:"default:0"`
	Date       string `gorm:"not null"` // YYYY-MM-DD
}

type RuntimeState struct {
	Key       string `gorm:"primaryKey"`
	Value     string
	UpdatedAt time.Time
}
