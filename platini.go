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
)

type route struct {
	method string
	path   string
	names  []string
	fv     reflect.Value
	ft     reflect.Type
	in     int
	out    int
}

type Mux struct {
	routes []route
}

var defaultMux Mux

func handler(w http.ResponseWriter, r *http.Request) {
	var rp *route
	for _, route := range defaultMux.routes {
		if route.method == r.Method {
			rp = &route
			break
		}
	}
	if rp == nil {
		return
	}
	values := make([]string, len(rp.names))
	match := false
	for i, p := range strings.Split(r.RequestURI, "/") {
		if p == "" {
			continue
		}
		if i >= len(rp.names) {
			break
		}
		if rp.names[i] != "" {
			values[i] = p
			match = true
		}
	}
	if !match {
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
	if len(rl) > 0 {
		res := make([]interface{}, len(rl))
		for i := 0; i < rp.out; i++ {
			v := rl[i].Interface()
			if err, ok := v.(error); ok {
				panic(err)
			}
			res[i] = v
		}
		if len(rl) == 1 {
			if res[0] != nil {
				json.NewEncoder(w).Encode(res[0])
			}
		} else {
			json.NewEncoder(w).Encode(res)
		}
	}
}

func HandleFunc(method, path string, fn interface{}) {
	names := strings.Split(path, "/")
	for i := 0; i < len(names); i++ {
		if re.MatchString(names[i]) {
			names[i] = names[i][1:]
		} else {
			names[i] = ""
		}
	}

	if len(defaultMux.routes) == 0 {
		http.HandleFunc("/", handler)
	}
	fv := reflect.ValueOf(fn)
	ft := fv.Type()
	defaultMux.routes = append(defaultMux.routes, route{
		method: method,
		path:   path,
		names:  names,
		fv:     fv,
		ft:     ft,
		in:     ft.NumIn(),
		out:    ft.NumOut(),
	})
}
