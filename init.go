package filter

import (
	"github.com/mholt/caddy"
	"github.com/mholt/caddy/caddyhttp/httpserver"
	"regexp"
	"strconv"
)

func init() {
	caddy.RegisterPlugin("filter", caddy.Plugin{
		ServerType: "http",
		Action:     setup,
	})
}

func setup(controller *caddy.Controller) error {
	handler, err := parseConfiguration(controller)
	if err != nil {
		return err
	}

	config := httpserver.GetConfig(controller)
	config.AddMiddleware(func(next httpserver.Handler) httpserver.Handler {
		handler.next = next
		return handler
	})

	return nil
}

func parseConfiguration(controller *caddy.Controller) (*filterHandler, error) {
	handler := new(filterHandler)
	handler.rules = []*rule{}
	handler.maximumBufferSize = defaultMaxBufferSize

	for controller.Next() {
		args := controller.RemainingArgs()
		if len(args) <= 0 {
			return nil, controller.Errf("No command provided.")
		}
		var err error
		switch args[0] {
		case "rule":
			err = evalRule(controller, args[1:], handler)
		case "maximumBufferSize":
			err = evalMaximumBufferSize(controller, args[1:], handler)
		default:
			err = controller.Errf("Unknown command '%v'.", args[0])
		}
		if err != nil {
			return nil, err
		}
	}
	return handler, nil
}

func evalRule(controller *caddy.Controller, args []string, target *filterHandler) (err error) {
	if len(args) > 0 {
		return controller.Errf("No more arguments for filter command 'rule' supported.")
	}
	targetRule := new(rule)
	for controller.NextBlock() {
		propertyName := controller.Val()
		switch propertyName {
		case "path":
			err = evalPath(controller, targetRule)
		case "content_type":
			err = evalContentType(controller, targetRule)
		case "search_pattern":
			err = evalSearchPattern(controller, targetRule)
		case "replacement":
			err = evalReplacement(controller, targetRule)
		default:
			err = controller.Errf("Unknown property name '%v'.", propertyName)
		}
		if err != nil {
			return err
		}
	}
	if targetRule.path == nil && targetRule.contentType == nil {
		return controller.Errf("Neither 'path' nor 'content_type' definition was provided for filter.")
	}
	if targetRule.searchPattern == nil {
		return controller.Errf("No 'search_pattern' definition was provided for filter.")
	}
	target.rules = append(target.rules, targetRule)
	return nil
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

func evalMaximumBufferSize(controller *caddy.Controller, args []string, target *filterHandler) (err error) {
	if len(args) != 1 {
		return controller.Errf("There are exact one argument for filter command 'maximumBufferSize' expected.")
	}
	value, err := strconv.Atoi(args[0])
	if err != nil {
		return controller.Errf("There is no valid value for filter command 'maximumBufferSize' provided. Got: %v", err)
	}
	target.maximumBufferSize = value
	return nil
}
