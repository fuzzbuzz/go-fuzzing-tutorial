package advancedtechniques

import (
	"regexp"
	"strings"
	"testing"

	"github.com/beego/beego/v2/server/web"
	context "github.com/beego/beego/v2/server/web/context"
)

func FuzzMatch(f *testing.F) {
	// Create some example routes for our rest API
	route1 := "/prefix/abc.html"
	route2 := "/1/2/3/hi.json"
	route3 := "/hel/lo/wo/rl/d"

	// Add them as valid routes to the router
	tree := web.NewTree()
	tree.AddRouter(route1, route1)
	tree.AddRouter(route2, route2)
	tree.AddRouter(route3, route3)

	// Add the routes as seeds for the fuzzer’s generation
	f.Add(route1)
	f.Add(route2)
	f.Add(route3)

	slashRE := regexp.MustCompile("//+")
	f.Fuzz(func(t *testing.T, pattern string) {
		// Filter repeated slashes, since beego fixes these up in an expected way
		// i.e. /prefix//abc.html matches /prefix/abc.html
		pattern = slashRE.ReplaceAllString(pattern, "/")

		// Try to match the fuzzed pattern to one of our defined API routes
		obj := tree.Match(pattern, context.NewContext())

		// Check if the pattern matches the example route
		if obj != nil {
			if matchedRoute := obj.(string); matchedRoute != "" {
				// Make sure the pattern and example route
				// share the same prefix
				if !strings.HasPrefix(pattern, matchedRoute) {
					t.Fatal("Found match with incorrect prefix", pattern, matchedRoute)
				}
			}
		}
	})
}
