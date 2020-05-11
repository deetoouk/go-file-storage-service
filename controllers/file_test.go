package controllers

import (
	"testing"

	"github.com/gin-gonic/gin"
)

// Get fetches a file by reference
func TestGet(t *testing.T) {
	hf = NewFileController()

	c = gin.CreateTestContext()

	hf.get(c)
}
