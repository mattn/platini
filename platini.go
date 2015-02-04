package platini

import (
	"encoding/json"
	"net/http"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

var (
	re = regexp.MustCompile(`^:[a-zA-Z][a-zA-Z0-9_]*$`)
	wt = reflect.TypeOf((*http.ResponseWriter)(nil)).Elem()
	rt = reflect.TypeOf((**http.Request)(nil)).Elem()
	et = reflect.TypeOf((*error)(nil)).Elem()
)

type route struct {
	method  string
	path    string
	names   []string
	fv      reflect.Value
	ft      reflect.Type
	in      int
	out     int
	params  int
	handler http.Handler
}

type Mux struct {
	routes []route
}

var DefaultMux = new(Mux)

func (m *Mux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var rp *route
	var values []string
	for _, route := range m.routes {
		if route.params == 0 && route.method == r.Method && r.RequestURI == route.path {
			rp = &route
			break
		}
	}
	if rp == nil {
	routeLoop:
		for _, route := range m.routes {
			if route.params > 0 && route.method == r.Method {
				match := false
				v := make([]string, len(route.names))
				for i, p := range strings.Split(r.RequestURI, "/") {
					if p == "" {
						continue
					}
					if i >= len(route.names) {
						break
					}
					if route.names[i] != "" {
						v[i] = p
						match = true
					}
				}
				if match {
					rp = &route
					values = v
					break routeLoop
				}
			}
		}
	}
	if rp == nil {
		for _, route := range m.routes {
			if route.method == "" && route.handler != nil {
				route.handler.ServeHTTP(w, r)
				return
			}
		}
	}
	if rp == nil {
		http.NotFound(w, r)
		return
	}
	args := make([]reflect.Value, rp.in)
	for i := 0; i < rp.in; i++ {
		it := rp.ft.In(i)
		if it.ConvertibleTo(wt) {
			args[i] = reflect.ValueOf(w).Convert(wt)
		} else if it.ConvertibleTo(rt) {
			args[i] = reflect.ValueOf(r).Convert(rt)
		} else {
			args[i] = reflect.New(it.Elem())
			at := args[i].Type()
			for at.Kind() == reflect.Ptr {
				at = at.Elem()
			}
			if at.Kind() == reflect.Struct {
				for f := 0; f < at.NumField(); f++ {
					tag := at.Field(f).Tag.Get("json")
					done := false
					for ni, name := range rp.names {
						if name == "" || tag != name {
							continue
						}
						field := args[i].Elem().Field(f)
						vv := reflect.ValueOf(values[ni])
						if vv.Type().ConvertibleTo(field.Type()) {
							field.Set(vv.Convert(field.Type()))
							done = true
						}
					}
					if !done {
						fnm := at.Field(f).Name
						for ni, name := range rp.names {
							if name == "" || strings.Title(name) != strings.Title(fnm) {
								continue
							}
							field := args[i].Elem().FieldByName(fnm)
							switch field.Kind() {
							case reflect.Int, reflect.Int16, reflect.Int32, reflect.Int64:
								if nv, err := strconv.Atoi(values[ni]); err == nil {
									field.Set(reflect.ValueOf(nv).Convert(field.Type()))
								}
							case reflect.String:
								field.Set(reflect.ValueOf(values[ni]))
							}
						}
					}
				}
			}
		}
	}
	rl := rp.fv.Call(args)
	if rp.out == 0 {
		return
	}
	if rp.ft.Out(rp.out-1) == et {
		if !rl[rp.out-1].IsNil() {
			http.Error(w, rl[rp.out-1].Interface().(error).Error(), http.StatusInternalServerError)
			return
		}
		rl = rl[:rp.out-1]
		rp.out--
	}
	if rp.out == 0 {
		return
	}
	res := []interface{}{}
	for i := 0; i < rp.out; i++ {
		res = append(res, rl[i].Interface())
	}
	if len(res) == 1 {
		if res[0] != nil {
			if s, ok := res[0].(string); ok {
				w.Write([]byte(s))
			} else if b, ok := res[0].([]byte); ok {
				w.Write(b)
			} else {
				json.NewEncoder(w).Encode(res[0])
			}
		}
	} else {
		json.NewEncoder(w).Encode(res)
	}
}

func (m *Mux) registerHandler(method, path string, fn interface{}) {
	names := strings.Split(path, "/")
	params := 0
	for i := 0; i < len(names); i++ {
		if re.MatchString(names[i]) {
			names[i] = names[i][1:]
			params++
		} else {
			names[i] = ""
		}
	}

	fv := reflect.ValueOf(fn)
	ft := fv.Type()
	m.routes = append(m.routes, route{
		method: method,
		path:   path,
		names:  names,
		fv:     fv,
		ft:     ft,
		in:     ft.NumIn(),
		out:    ft.NumOut(),
		params: params,
	})
}

func (m *Mux) Get(path string, fn interface{}) {
	m.registerHandler("GET", path, fn)
}

func (m *Mux) Post(path string, fn interface{}) {
	m.registerHandler("POST", path, fn)
}

func (m *Mux) Handle(path string, handler http.Handler) {
	m.routes = append(m.routes, route{
		path:    path,
		handler: handler,
	})
}

func Get(path string, fn interface{}) {
	DefaultMux.Get(path, fn)
}

func Post(path string, fn interface{}) {
	DefaultMux.Post(path, fn)
}

func Handle(path string, handler http.Handler) {
	DefaultMux.Handle(path, handler)
}
