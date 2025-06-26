package routing

import "net/http"

// HandlerFunc is the signature for route handlers
type HandlerFunc func(w http.ResponseWriter, r *http.Request)

// MiddlewareFunc wraps a handler with additional functionality
type MiddlewareFunc func(HandlerFunc) HandlerFunc

// Route represents a single route configuration
type Route struct {
	method      string
	pattern     string
	handler     HandlerFunc
	middlewares []MiddlewareFunc
}

// handle returns the final handler with all middlewares applied
func (r *Route) handle() HandlerFunc {
	h := r.handler
	// Apply middlewares in reverse order
	for i := len(r.middlewares) - 1; i >= 0; i-- {
		h = r.middlewares[i](h)
	}
	return h
}

// RouteGroup represents a group of routes with common prefix/middleware
type RouteGroup struct {
	prefix      string
	middlewares []MiddlewareFunc
	routes      []*Route
	groups      []*RouteGroup
}

// flatten returns all routes in this group and subgroups with applied prefix and middleware
func (g *RouteGroup) flatten(parentPrefix string, parentMiddlewares []MiddlewareFunc) []*Route {
	var result []*Route

	fullPrefix := parentPrefix + g.prefix
	allMiddlewares := append(parentMiddlewares, g.middlewares...)

	// Add direct routes
	for _, route := range g.routes {
		r := &Route{
			method:      route.method,
			pattern:     fullPrefix + route.pattern,
			handler:     route.handler,
			middlewares: append(allMiddlewares, route.middlewares...),
		}
		result = append(result, r)
	}

	// Recursively add routes from subgroups
	for _, subgroup := range g.groups {
		result = append(result, subgroup.flatten(fullPrefix, allMiddlewares)...)
	}

	return result
}

// RouteBuilder provides route construction methods
type RouteBuilder struct{}

func NewRoute() *RouteBuilder {
	return &RouteBuilder{}
}

// GET creates a GET route
func (rb *RouteBuilder) GET(pattern string, handler HandlerFunc) *Route {
	return &Route{method: "GET", pattern: pattern, handler: handler}
}

// POST creates a POST route
func (rb *RouteBuilder) POST(pattern string, handler HandlerFunc) *Route {
	return &Route{method: "POST", pattern: pattern, handler: handler}
}

// PUT creates a PUT route
func (rb *RouteBuilder) PUT(pattern string, handler HandlerFunc) *Route {
	return &Route{method: "PUT", pattern: pattern, handler: handler}
}

// DELETE creates a DELETE route
func (rb *RouteBuilder) DELETE(pattern string, handler HandlerFunc) *Route {
	return &Route{method: "DELETE", pattern: pattern, handler: handler}
}

// PATCH creates a PATCH route
func (rb *RouteBuilder) PATCH(pattern string, handler HandlerFunc) *Route {
	return &Route{method: "PATCH", pattern: pattern, handler: handler}
}

// Group creates a route group
func (rb *RouteBuilder) Group(prefix string, items ...any) *RouteGroup {
	group := &RouteGroup{prefix: prefix}

	for _, item := range items {
		switch v := item.(type) {
		case *Route:
			group.routes = append(group.routes, v)
		case *RouteGroup:
			group.groups = append(group.groups, v)
		case []MiddlewareFunc:
			group.middlewares = append(group.middlewares, v...)
		case MiddlewareFunc:
			group.middlewares = append(group.middlewares, v)
		}
	}

	return group
}

// App represents the Bold application
type NetHTTPApp struct {
	routes      []*Route
	groups      []*RouteGroup
	middlewares []MiddlewareFunc
}

// Routes configures the application routes
func (app *NetHTTPApp) Routes(items ...any) {
	for _, item := range items {
		switch v := item.(type) {
		case *Route:
			app.routes = append(app.routes, v)
		case *RouteGroup:
			app.groups = append(app.groups, v)
		}
	}
}

// Handler returns an http.Handler for the application
func (app *NetHTTPApp) Handler() http.Handler {
	mux := http.NewServeMux()

	// Collect all routes
	allRoutes := make([]*Route, 0)

	// Add direct routes
	allRoutes = append(allRoutes, app.routes...)

	// Add routes from groups
	for _, group := range app.groups {
		allRoutes = append(allRoutes, group.flatten("", nil)...)
	}

	// Register routes with mux
	for _, route := range allRoutes {
		pattern := route.method + " " + route.pattern
		handler := route.handle()

		// Apply global middlewares
		for i := len(app.middlewares) - 1; i >= 0; i-- {
			handler = app.middlewares[i](handler)
		}

		mux.HandleFunc(pattern, handler)
	}

	return mux
}

// Listen starts the HTTP server
func (app *NetHTTPApp) Listen(addr string) error {
	return http.ListenAndServe(addr, app.Handler())
}
