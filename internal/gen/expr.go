package main

import (
	"bytes"
	"go/format"
	"io/ioutil"
	"reflect"
	"strings"
	"text/template"

	"github.com/hekmon/transmissionrpc"
)

var exprGenTemplate = template.Must(template.New("exprGen").Parse(`
package jobs

// *********************************************
// Code generated by gen/expr.go -- DO NOT EDIT.
// *********************************************

import (
	"time"

	"github.com/hekmon/cunits/v2"
	"github.com/hekmon/transmissionrpc"
)

// TransmissionTorrent is a generated, pointer-free, and less safe variant of transmissionrpc.Torrent to make using the 
// expr package easier.
type TransmissionTorrent struct {
	{{- range .Props }}
	{{ .FieldName }} {{.FieldType }}
	{{- end }}

	*StoredTorrentInfo

	// for internal, ephemeral use
	sonarrDropPaths map[string]bool
}

// ToTransmissionTorrent converts the library struct to our generated struct.
func ToTransmissionTorrent(input transmissionrpc.Torrent, sonarrDropPaths map[string]bool) TransmissionTorrent {
	return TransmissionTorrent{
		{{- range .Props }}
		{{ .FieldName }}: {{ if .Dereference }}*{{ end }}input.{{ .FieldName }},
		{{- end }}
		sonarrDropPaths: sonarrDropPaths,
	}
}
`))

type exprGenInput struct {
	Props []exprGenInputProps
}

type exprGenInputProps struct {
	FieldName,
	FieldType string
	Dereference bool
}

func main() {
	// create jobs/expr_gen.go
	torrentType := reflect.TypeOf(transmissionrpc.Torrent{})
	props := make([]exprGenInputProps, torrentType.NumField())
	for i := 0; i < torrentType.NumField(); i++ {
		var dereferenceRequired bool
		field := torrentType.Field(i)
		fieldType := field.Type.String()
		// this is so dumb
		dereferenceRequired = strings.HasPrefix(fieldType, "*")
		if dereferenceRequired {
			fieldType = strings.TrimLeft(fieldType, "*")
		}
		props[i] = exprGenInputProps{
			FieldName:   field.Name,
			FieldType:   fieldType,
			Dereference: dereferenceRequired,
		}
	}
	// format + save
	toFormat := &bytes.Buffer{}
	err := exprGenTemplate.Execute(toFormat, exprGenInput{props})
	if err != nil {
		panic(err)
	}
	formatted, err := format.Source(toFormat.Bytes())
	if err != nil {
		panic(err)
	}
	err = ioutil.WriteFile("jobs/expr_gen.go", formatted, 0644)
	if err != nil {
		panic(err)
	}
}
