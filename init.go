package filter

import (
	"fmt"
	"github.com/mholt/caddy"
	"github.com/mholt/caddy/caddyhttp/httpserver"
	"regexp"
)

func init() {
	caddy.RegisterPlugin("filter", caddy.Plugin{
		ServerType: "http",
		Action:     setup,
	})
}

func setup(controller *caddy.Controller) error {
	rules, err := parseRules(controller)
	if err != nil {
		return err
	}

	config := httpserver.GetConfig(controller)
	mid := func(next httpserver.Handler) httpserver.Handler {
		return filterHandler{Next: next, Rules: rules}
	}
	config.AddMiddleware(mid)

	return nil
}

func parseRules(controller *caddy.Controller) ([]*rule, error) {
	rules := []*rule{}

	for controller.Next() {
		file := controller.File()
		line := controller.Line()
		target := new(rule)
		for controller.NextBlock() {
			propertyName := controller.Val()
			switch propertyName {
			case "path":
				evalPath(controller, target)
			case "content_type":
				evalContentType(controller, target)
			case "search_pattern":
				evalSearchPattern(controller, target)
			case "replacement":
				evalReplacement(controller, target)
			}
		}
		if target.path == nil && target.contentType == nil {
			return rules, fmt.Errorf("%s:%d - neither 'path' nor 'content_type' definition was provided for filter.", file, line)
		}
		if target.searchPattern == nil {
			return rules, fmt.Errorf("%s:%d - no 'search_pattern' definition was provided for filter.", file, line)
		}
		rules = append(rules, target)
	}
	return rules, nil
}

func evalPath(controller *caddy.Controller, target *rule) error {
	return evalRegexpProperty(controller, func(value *regexp.Regexp) error {
		target.path = value
		return nil
	})
}

func evalContentType(controller *caddy.Controller, target *rule) error {
	return evalRegexpProperty(controller, func(value *regexp.Regexp) error {
		target.contentType = value
		return nil
	})
}

func evalSearchPattern(controller *caddy.Controller, target *rule) error {
	return evalRegexpProperty(controller, func(value *regexp.Regexp) error {
		target.searchPattern = value
		return nil
	})
}

func evalReplacement(controller *caddy.Controller, target *rule) error {
	return evalSimpleProperty(controller, func(value string) error {
		target.replacement = []byte(value)
		return nil
	})
}

func evalSimpleProperty(controller *caddy.Controller, setter func(string) error) error {
	args := controller.RemainingArgs()
	if len(args) != 1 {
		return controller.ArgErr()
	}
	return setter(args[0])
}

func evalRegexpProperty(controller *caddy.Controller, setter func(*regexp.Regexp) error) error {
	return evalSimpleProperty(controller, func(plainValue string) error {
		value, err := regexp.Compile(plainValue)
		if err != nil {
			return err
		}
		return setter(value)
	})
}
