package gee

import (
	"log"
	"net/http"
	"strings"
)

type router struct {
	roots    map[string]*node // key part val node
	handlers map[string]HandlerFunc
}

func newRouter() *router {
	return &router{
		roots:    make(map[string]*node),
		handlers: make(map[string]HandlerFunc),
	}
}

// only single * allowed
func parsePattern(pattern string) []string {
	vs := strings.Split(pattern, "/")
	parts := make([]string, 0)
	for _, part := range vs {
		if part == "" {
			continue
		}
		parts = append(parts, part)
		if part[0] == '*' {
			break
		}
	}
	return parts
}

func (r *router) addRoute(method string, pattern string, handler HandlerFunc) {
	log.Printf("Route %4s - %s", method, pattern)

	parts := parsePattern(pattern)
	key := method + "-" + pattern

	root := r.roots[method]
	if root == nil {
		r.roots[method] = &node{}
	}
	r.roots[method].insert(pattern, parts, 0)
	r.handlers[key] = handler
}

// return node and params
func (r *router) getRoute(method string, path string) (*node, map[string]string) {
	root, ok := r.roots[method]
	if !ok {
		return nil, nil
	}

	// find node
	realParts := parsePattern(path)
	findNode := root.search(realParts, 0)

	// generalize the param
	if findNode != nil {
		// key param value paramVal
		params := make(map[string]string)
		findNodeParts := parsePattern(findNode.pattern)
		for index, part := range findNodeParts {
			if part[0] == ':' {
				paramKey := part[1:]
				params[paramKey] = realParts[index]
			}
			if part[0] == '*' && len(part) > 1 {
				params[part[1:]] = strings.Join(realParts[index:], "/")
				break
			}
		}
		return findNode, params
	}
	return nil, nil
}

func (r *router) handle(c *Context) {
	n, params := r.getRoute(c.Method, c.Path)
	if n != nil {
		c.Params = params
		key := c.Method + "-" + n.pattern
		r.handlers[key](c)
	} else {
		c.String(http.StatusNotFound, "404 NOT FOUND: %s\n", c.Path)
	}
}
