package webgo

import (
	"net/http"
	"testing"
)

func TestRouterMatch(t *testing.T) {

	newRouteOption := func() *RouteOptions {
		return &RouteOptions{
			Controller:  new(TestController),
			Action:      "Invoke",
			ContentType: CT_JSON,
		}
	}

	for _, td := range []struct {
		SrcMethod string
		SrcUrl    string
		ReqMethod string
		ReqUrl    string
		Res       *Match
	}{
		{
			SrcMethod: http.MethodPost,
			SrcUrl:    "/c",
			ReqMethod: http.MethodPost,
			ReqUrl:    "/b",
			Res:       nil,
		},
		{
			SrcMethod: http.MethodPost,
			SrcUrl:    "/c",
			ReqMethod: http.MethodGet,
			ReqUrl:    "/c",
			Res:       nil,
		},
		{
			SrcMethod: http.MethodPost,
			SrcUrl:    "/a",
			ReqMethod: http.MethodPost,
			ReqUrl:    "/a",
			Res:       &Match{Params: Params{}},
		},
		{
			SrcMethod: http.MethodPost,
			SrcUrl:    "/a/:v1/:v2",
			ReqMethod: http.MethodPost,
			ReqUrl:    "/a/1/2",
			Res:       &Match{Params: Params{"v1": "1", "v2": "2"}},
		},
		{
			SrcMethod: http.MethodPost,
			SrcUrl:    "/a/:v1/:v2",
			ReqMethod: http.MethodPost,
			ReqUrl:    "/a/1",
			Res:       nil,
		},
	} {

		router := new(Router)
		if err := router.Add(td.SrcMethod, td.SrcUrl, newRouteOption()); err != nil {
			t.Fatal(err, td)
		}

		match := router.Match(td.ReqMethod, td.ReqUrl)

		if td.Res == nil && match != nil {
			t.Error("Fail:", td.SrcMethod, td.SrcUrl, td.ReqMethod, td.ReqUrl)
		} else if td.Res != nil && match == nil {
			t.Error("Fail:", td.SrcMethod, td.SrcUrl, td.ReqMethod, td.ReqUrl)
		} else {

			if td.Res == nil && td.Res == match {
				continue // ok
			}

			if len(match.Params) != len(td.Res.Params) {
				t.Errorf("Fail: '%s' %v !- %v", td.SrcMethod, match.Params, td.Res.Params)
			} else {
				for k, srcVal := range td.Res.Params {
					if resVal, ok := match.Params[k]; !ok || srcVal != resVal {
						t.Errorf("Fail: '%s' %v !- %v", td.SrcMethod, match.Params, td.Res.Params)
					}
				}
			}
		}

	}
}

func TestRouterCopy(t *testing.T) {

	type RouteDesc struct {
		Method string
		Path   string
	}

	createRoute := func(args ...*RouteDesc) *Router {
		retval := new(Router)
		for _, item := range args {
			retval.Add(
				item.Method,
				item.Path,
				&RouteOptions{
					Controller:  new(TestController),
					Action:      "Invoke",
					ContentType: CT_JSON,
				},
			)
		}

		return retval
	}

	for _, td := range []struct {
		Src *Router
		Dst *Router
		Res *Router
		Err string
	}{
		{
			Src: createRoute(&RouteDesc{http.MethodPost, "/a"}),
			Dst: createRoute(&RouteDesc{http.MethodPost, "/b"}),
			Res: createRoute(&RouteDesc{http.MethodPost, "/a"}, &RouteDesc{http.MethodPost, "/b"}),
			Err: "",
		},
		{
			Src: createRoute(&RouteDesc{http.MethodPost, "/a"}),
			Dst: createRoute(),
			Res: createRoute(&RouteDesc{http.MethodPost, "/a"}),
			Err: "",
		},
		{
			Src: createRoute(&RouteDesc{http.MethodPost, "/a"}),
			Dst: createRoute(&RouteDesc{http.MethodPost, "/a"}),
			Res: createRoute(&RouteDesc{http.MethodPost, "/a"}),
			Err: "Route path already use: 'POST'->'/a'.",
		},
	} {

		err := td.Src.Copy(td.Dst)
		switch td.Err {
		case "":
			if err != nil {
				t.Fatalf("Fail: %s %+v", err, td)
			}

			if len(td.Dst.routes) == 0 || len(td.Dst.routes) != len(td.Res.routes) {
				t.Error("Not equal count: %d != %d", len(td.Dst.routes), len(td.Res.routes))
			} else {

				for method, resRouting := range td.Res.routes {
					dstRouting, ok := td.Dst.routes[method]
					if !ok {
						t.Errorf("Not found routes for method: '%s'", method)
						continue
					}

					for routePath, resRouteDesc := range resRouting {
						dstRouteDesc, ok := dstRouting[routePath]
						if !ok {
							t.Errorf("Not found route: '%s'", routePath)
							continue
						}

						if resRouteDesc.Pattern != dstRouteDesc.Pattern {
							t.Errorf("Wrong description '%s:%s': '%s' != '%s'", method, routePath, resRouteDesc.Pattern, dstRouteDesc.Pattern)
						}
					}
				}
			}

		default:
			if err == nil {
				t.Fatalf("Fail: %v", td)
			}

			if err.Error() != td.Err {
				t.Errorf("Fail: '%s'!='%s'", err, td.Err)
			}
		}
	}
}
