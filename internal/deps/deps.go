// Package deps imports all project dependencies to ensure they are tracked in go.mod
// This file will be removed once actual implementation uses these packages
package deps

import (
	_ "github.com/gin-gonic/gin"
	_ "github.com/golang-jwt/jwt/v5"
	_ "github.com/gorilla/websocket"
	_ "golang.org/x/crypto/bcrypt"
	_ "gorm.io/driver/sqlite"
	_ "gorm.io/gorm"
)
