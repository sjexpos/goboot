package web

import "github.com/gin-gonic/gin"

type Middleware interface {
	//gin.HandlerFunc
	DoFilter(*gin.Context)
}
