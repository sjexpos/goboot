package gorm

import (
	"testing"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func Test(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("gorm.db"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Gorm connection error: %v", err)
	}
	f := CreateOpenSessionInViewFilter(db)
	ctx := &gin.Context{}
	f(ctx)
}
