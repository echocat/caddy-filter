package filter

import (
	"errors"
	"fmt"
	"github.com/mholt/caddy"
	"github.com/mholt/caddy/caddyhttp/httpserver"
	. "gopkg.in/check.v1"
	"regexp"
	"regexp/syntax"
)

type initTest struct{}

func init() {
	Suite(&initTest{})
}

func (s *initTest) Test_setup(c *C) {
	controller := s.newControllerFor("filter rule {\npath myPath\ncontent_type myContentType\nsearch_pattern mySearchPattern\nreplacement myReplacement\n}\n")
	err := setup(controller)
	c.Assert(err, IsNil)
	config := httpserver.GetConfig(controller)
	middlewares := config.Middleware()
	c.Assert(len(middlewares), Equals, 1)
	handler, ok := middlewares[0](newMockHandler("moo", 200)).(*filterHandler)
	c.Assert(ok, Equals, true)
	c.Assert(len(handler.rules), Equals, 1)
	r := handler.rules[0]
	c.Assert(r.path.String(), Equals, "myPath")
	c.Assert(r.contentType.String(), Equals, "myContentType")
	c.Assert(r.searchPattern.String(), Equals, "mySearchPattern")
	c.Assert(string(r.replacement), Equals, "myReplacement")
}

func (s *initTest) Test_parseConfiguration_default(c *C) {
	handler, err := parseConfiguration(s.newControllerFor("filter {\nrule {\npath myPath\ncontent_type myContentType\nsearch_pattern mySearchPattern\nreplacement myReplacement\n}\n}\n"))
	c.Assert(err, IsNil)
	c.Assert(len(handler.rules), Equals, 1)
	r := handler.rules[0]
	c.Assert(r.path.String(), Equals, "myPath")
	c.Assert(r.contentType.String(), Equals, "myContentType")
	c.Assert(r.searchPattern.String(), Equals, "mySearchPattern")
	c.Assert(string(r.replacement), Equals, "myReplacement")

	fmt.Println()
	fmt.Println()
	fmt.Println()

	handler, err = parseConfiguration(s.newControllerFor("filter {\n" +
		"rule {\npath myPath\nsearch_pattern mySearchPattern\n}\n" +
		"rule {\npath myPath2\nsearch_pattern mySearchPattern2\n}\n" +
		"max_buffer_size 666\n" +
		"}",
	),
	)
	c.Assert(err, IsNil)
	c.Assert(len(handler.rules), Equals, 2)
	r = handler.rules[0]
	c.Assert(r.path.String(), Equals, "myPath")
	c.Assert(r.searchPattern.String(), Equals, "mySearchPattern")
	r = handler.rules[1]
	c.Assert(r.path.String(), Equals, "myPath2")
	c.Assert(r.searchPattern.String(), Equals, "mySearchPattern2")
	c.Assert(handler.maximumBufferSize, Equals, 666)

	_, err = parseConfiguration(s.newControllerFor("filter moo"))
	c.Assert(err, DeepEquals, errors.New("Testfile:1 - Parse error: Unknown directive: moo"))

	_, err = parseConfiguration(s.newControllerFor("filter"))
	c.Assert(err, DeepEquals, errors.New("Testfile:1 - Parse error: No rule block provided."))
}

func (s *initTest) Test_parseConfiguration_directNamed(c *C) {
	handler, err := parseConfiguration(s.newControllerFor("filter rule {\npath myPath\ncontent_type myContentType\nsearch_pattern mySearchPattern\nreplacement myReplacement\n}\n"))
	c.Assert(err, IsNil)
	c.Assert(len(handler.rules), Equals, 1)
	r := handler.rules[0]
	c.Assert(r.path.String(), Equals, "myPath")
	c.Assert(r.contentType.String(), Equals, "myContentType")
	c.Assert(r.searchPattern.String(), Equals, "mySearchPattern")
	c.Assert(string(r.replacement), Equals, "myReplacement")

	handler, err = parseConfiguration(s.newControllerFor(
		"filter rule {\npath myPath\nsearch_pattern mySearchPattern\n}\n" +
			"filter rule {\npath myPath2\nsearch_pattern mySearchPattern2\n}\n" +
			"filter max_buffer_size 666\n"),
	)
	c.Assert(err, IsNil)
	c.Assert(len(handler.rules), Equals, 2)
	r = handler.rules[0]
	c.Assert(r.path.String(), Equals, "myPath")
	c.Assert(r.searchPattern.String(), Equals, "mySearchPattern")
	r = handler.rules[1]
	c.Assert(r.path.String(), Equals, "myPath2")
	c.Assert(r.searchPattern.String(), Equals, "mySearchPattern2")
	c.Assert(handler.maximumBufferSize, Equals, 666)
}

func (s *initTest) Test_parseConfiguration_withReplacementFromFile(c *C) {
	handler, err := parseConfiguration(s.newControllerFor("filter rule {\npath myPath\ncontent_type myContentType\nsearch_pattern mySearchPattern\nreplacement @resources/test/testReplacement\n}\n"))
	c.Assert(err, IsNil)
	c.Assert(len(handler.rules), Equals, 1)
	r := handler.rules[0]
	c.Assert(r.path.String(), Equals, "myPath")
	c.Assert(r.contentType.String(), Equals, "myContentType")
	c.Assert(r.searchPattern.String(), Equals, "mySearchPattern")
	c.Assert(string(r.replacement), Equals, "Replacement from file.\n")
}

func (s *initTest) Test_evalSimpleOption(c *C) {
	err := evalSimpleOption(s.newControllerFor("\"my value\""), func(value string) error {
		c.Assert(value, Equals, "my value")
		return nil
	})
	c.Assert(err, IsNil)

	err = evalSimpleOption(s.newControllerFor(""), func(value string) error {
		c.Error("This method should not be called.")
		return nil
	})
	c.Assert(err, DeepEquals, errors.New("Testfile:1 - Parse error: Wrong argument count or unexpected line ending after 'start'"))
}

func (s *initTest) Test_evalRegexpOption(c *C) {
	err := evalRegexpOption(s.newControllerFor("f.*bar"), func(value *regexp.Regexp) error {
		c.Assert(value.MatchString("foobar"), Equals, true)
		return nil
	})
	c.Assert(err, IsNil)

	err = evalRegexpOption(s.newControllerFor("<???"), func(value *regexp.Regexp) error {
		c.Error("This method should not be called.")
		return nil
	})
	c.Assert(err, DeepEquals, &syntax.Error{Code: "invalid nested repetition operator", Expr: "???"})
}

func (s *initTest) Test_evalPath(c *C) {
	r := new(rule)
	err := evalPath(s.newControllerFor("f.*bar"), r)
	c.Assert(err, IsNil)
	c.Assert(r.path.String(), Equals, "f.*bar")
}

func (s *initTest) Test_evalContentType(c *C) {
	r := new(rule)
	err := evalContentType(s.newControllerFor("f.*bar"), r)
	c.Assert(err, IsNil)
	c.Assert(r.contentType.String(), Equals, "f.*bar")
}

func (s *initTest) Test_searchPattern(c *C) {
	r := new(rule)
	err := evalSearchPattern(s.newControllerFor("f.*bar"), r)
	c.Assert(err, IsNil)
	c.Assert(r.searchPattern.String(), Equals, "f.*bar")
}

func (s *initTest) Test_evalReplacement(c *C) {
	r := new(rule)
	err := evalReplacement(s.newControllerFor("foobar"), r)
	c.Assert(err, IsNil)
	c.Assert(string(r.replacement), Equals, "foobar")
}

func (s *initTest) Test_evalRule(c *C) {
	handler := new(filterHandler)
	err := evalRule(s.newControllerFor("{\npath myPath\ncontent_type myContentType\nsearch_pattern mySearchPattern\nreplacement myReplacement\n}\n"), []string{}, handler)
	c.Assert(err, IsNil)
	c.Assert(len(handler.rules), Equals, 1)
	r := handler.rules[0]
	c.Assert(r.path.String(), Equals, "myPath")
	c.Assert(r.contentType.String(), Equals, "myContentType")
	c.Assert(r.searchPattern.String(), Equals, "mySearchPattern")
	c.Assert(string(r.replacement), Equals, "myReplacement")

	err = evalRule(s.newControllerFor("{\nfoo bar\n}\n"), []string{}, handler)
	c.Assert(err, DeepEquals, errors.New("Testfile:2 - Parse error: Unknown option: foo"))

	err = evalRule(s.newControllerFor("{\n}\n"), []string{}, handler)
	c.Assert(err, DeepEquals, errors.New("Testfile:2 - Parse error: Neither 'path' nor 'content_type' definition was provided for filter rule block."))

	err = evalRule(s.newControllerFor("{\npath myPath\n}\n"), []string{}, handler)
	c.Assert(err, DeepEquals, errors.New("Testfile:3 - Parse error: No 'search_pattern' definition was provided for filter rule block."))

	err = evalRule(s.newControllerFor(""), []string{"foo"}, handler)
	c.Assert(err, DeepEquals, errors.New("Testfile:1 - Parse error: No more arguments for filter block 'rule' supported."))
}

func (s *initTest) Test_evalMaximumBufferSize(c *C) {
	handler := new(filterHandler)
	err := evalMaximumBufferSize(s.newControllerFor(""), []string{"123"}, handler)
	c.Assert(err, IsNil)
	c.Assert(handler.maximumBufferSize, Equals, 123)

	err = evalMaximumBufferSize(s.newControllerFor(""), []string{}, handler)
	c.Assert(err, DeepEquals, errors.New("Testfile:1 - Parse error: There are exact one argument for filter directive 'max_buffer_size' expected."))

	err = evalMaximumBufferSize(s.newControllerFor(""), []string{"abc"}, handler)
	c.Assert(err, DeepEquals, errors.New("Testfile:1 - Parse error: There is no valid value for filter directive 'max_buffer_size' provided. Got: strconv.ParseInt: parsing \"abc\": invalid syntax"))
}

func (s *initTest) newControllerFor(plainTokens string) *caddy.Controller {
	controller := caddy.NewTestController("http", "start "+plainTokens)
	if !controller.Next() {
		panic("There must be an entry.")
	}
	return controller
}
