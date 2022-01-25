package template

import (
	"bytes"
	"strings"
	"text/template"
)

var httpTemplate = `
{{$svrType := .ServiceType}}
{{$svrName := .ServiceName}}
type {{.ServiceType}} interface {
{{- range .MethodSets}}
	{{.Name}}(context.Context, *{{.Request}}) (*{{.Reply}}, error)
{{- end}}
}

func Register{{.ServiceType}}Router(s *http.Server, srv {{.ServiceType}}) {
	r := s.Router()
	{{- range .Methods}}
	r.{{.Method}}("{{.Path}}", {{.Name}}Handler(srv))
	{{- end}}
}

{{range .Methods}}
func {{.Name}}Handler(srv {{$svrType}}) func(c *gin.Context) {
	return func(c *gin.Context) {
		var in {{.Request}}
		if err := c.Bind(&in); err != nil {
			c.AbortWithError(500, err)
			return
		}
		resp, err := srv.{{.Name}}(c.Copy(), &in)
		if err != nil {
			c.AbortWithError(500, err)
			return
		}
		http.JSON(resp, c)
	}
}
{{end}}
`

type ServiceDesc struct {
	ServiceType string // Greeter
	ServiceName string // helloworld.Greeter
	Metadata    string // api/helloworld/helloworld.proto
	Methods     []*MethodDesc
	MethodSets  map[string]*MethodDesc
}

type MethodDesc struct {
	// method
	Name    string
	Num     int
	Request string
	Reply   string
	// http_rule
	Path         string
	Method       string
	HasVars      bool
	HasBody      bool
	Body         string
	ResponseBody string
}

func (s *ServiceDesc) Execute() string {
	s.MethodSets = make(map[string]*MethodDesc)
	for _, m := range s.Methods {
		s.MethodSets[m.Name] = m
	}
	buf := new(bytes.Buffer)
	tmpl, err := template.New("http").Parse(strings.TrimSpace(httpTemplate))
	if err != nil {
		panic(err)
	}
	if err := tmpl.Execute(buf, s); err != nil {
		panic(err)
	}
	return strings.Trim(string(buf.Bytes()), "\r\n")
}
