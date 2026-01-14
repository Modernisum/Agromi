package router

import "github.com/gin-gonic/gin"

// RouteRegistrar defines a function that takes the Gin engine and adds routes to it
type RouteRegistrar func(*gin.Engine)

// Registry holds all the registered route functions
var Registry []RouteRegistrar

// Register adds a new route function to the Registry.
// This is called by 'init()' functions in feature files.
func Register(r RouteRegistrar) {
	Registry = append(Registry, r)
}
