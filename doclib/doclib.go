package doclib

import (
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3gen"
	orderedmap "github.com/wk8/go-ordered-map/v2"
)

type SetupData struct {
	URL             string
	ErrorStruct     any
	Info            Info
	errorStructName string
}

var (
	DocsSetupData *SetupData
	stringType    = openapi3.Types([]string{"string"})
)

func Setup() {
	if DocsSetupData == nil {
		panic("DocsSetupData is nil")
	}

	var err error

	badRequestSchema, err = openapi3gen.NewSchemaRefForValue(DocsSetupData.ErrorStruct, nil, SchemaInject(DocsSetupData.ErrorStruct))

	if err != nil {
		panic(err)
	}

	// Set errorSchemaName
	errorSchemaName := reflect.TypeOf(DocsSetupData.ErrorStruct).String()
	errorSchemaName = strings.ReplaceAll(errorSchemaName, "docs.", "")

	DocsSetupData.errorStructName = errorSchemaName

	IdSchema, err = openapi3gen.NewSchemaRefForValue("1234567890", nil)

	if err != nil {
		panic(err)
	}

	BoolSchema, err = openapi3gen.NewSchemaRefForValue(true, nil)

	if err != nil {
		panic(err)
	}

	api.Components.Schemas[DocsSetupData.errorStructName] = badRequestSchema

	api.Info = DocsSetupData.Info
	api.Servers[0].URL = DocsSetupData.URL
	api.Paths = orderedmap.New[string, Path]()
	api.Webhooks = orderedmap.New[string, Path]()
}

var api = Openapi{
	OpenAPI: "3.1.0",
	Servers: []Server{
		{
			Description: "Bytepurr API",
			Variables:   map[string]any{},
		},
	},
	Components: Component{
		Schemas:       make(map[string]any),
		Security:      make(map[string]Security),
		RequestBodies: make(map[string]ReqBody),
	},
}

var badRequestSchema *openapi3.SchemaRef

var IdSchema *openapi3.SchemaRef
var BoolSchema *openapi3.SchemaRef

func AddTag(name, description string) {
	api.Tags = append(api.Tags, Tag{
		Name:        name,
		Description: description,
	})
}

func AddSecuritySchema(id, header, description string) {
	api.Components.Security[id] = Security{
		Type:        "apiKey",
		Name:        header,
		In:          "header",
		Description: description,
	}
}

func SchemaInject(s any) openapi3gen.Option {
	return openapi3gen.SchemaCustomizer(func(name string, ft reflect.Type, tag reflect.StructTag, schema *openapi3.Schema) error {
		if tag.Get("description") != "" {
			schema.Description = tag.Get("description")
		}

		if tag.Get("dynexample") == "true" {
			var fname string
			for _, field := range reflect.VisibleFields(reflect.TypeOf(s)) {
				if field.Tag.Get("json") == name {
					fname = field.Name
				}
			}

			defaultVal := reflect.ValueOf(s).FieldByName(fname).Interface()
			schema.Example = defaultVal
		}

		if tag.Get("dynschema") == "true" {
			// Get schema data from the field
			var fname string
			for _, field := range reflect.VisibleFields(reflect.TypeOf(s)) {
				if field.Tag.Get("json") == name {
					fname = field.Name
				}
			}

			schemaData := reflect.ValueOf(s).FieldByName(fname).Interface()

			// Generate schema
			schemaRef, err := openapi3gen.NewSchemaRefForValue(schemaData, nil, SchemaInject(schemaData))

			if err != nil {
				panic(err)
			}

			schema.Properties = schemaRef.Value.Properties
		}

		if tag.Get("enum") != "" {
			// Split by comma
			enumVals := strings.Split(tag.Get("enum"), ",")

			schema.Enum = []any{}

			for _, val := range enumVals {
				schema.Enum = append(schema.Enum, val)
			}
		}

		if tag.Get("validate") != "" {
			// Split by comma
			validateVals := strings.Split(tag.Get("validate"), ",")

			for _, val := range validateVals {
				key := strings.Split(val, "=")[0]
				switch key {
				case "required":
					schema.Nullable = false
				case "oneof":
					enumVals := strings.Split(val, "=")[1]

					var enum []any

					for _, val := range strings.Split(enumVals, " ") {
						enum = append(enum, val)
					}

					schema.Enum = enum
				}
			}
		}

		switch ft.Name() {
		case "Text":
			schema.Type = &stringType
			schema.Nullable = true
		case "Int4":
			panic("never use pgtype/whatever.Int4, use int32 instead")
		case "Int8":
			panic("never use pgtype/whatever.Int8, use int/int64 instead")
		case "Timestamp":
			schema.Type = &stringType
			schema.Format = "date-time"
		case "Timestamptz":
			schema.Type = &stringType
			schema.Format = "date-time"
		case "Date":
			schema.Type = &stringType
			schema.Format = "date"
		case "Bool":
			panic("never use pgtype/whatever.Bool, use bool instead")
		case "UUID":
			schema.Type = &stringType
			schema.Format = "uuid"
		}

		if tag.Get("type") != "" {
			typ := openapi3.Types([]string{tag.Get("type")})
			schema.Type = &typ
		}

		return nil
	})
}

func Route(doc *Doc) {
	// Generate schemaName, taking out bad things

	// Basic checks
	if len(doc.Params) == 0 {
		doc.Params = []Parameter{}
	}

	if len(doc.AuthType) == 0 {
		doc.AuthType = []string{}
	}

	if len(doc.Tags) == 0 {
		panic("no tags set in route: " + doc.Pattern)
	}

	if len(doc.Params) > 0 {
		for _, param := range doc.Params {
			if param.In == "" {
				panic("no in set in route: " + doc.Pattern)
			}

			if param.Name == "" {
				panic("no name set in route: " + doc.Pattern)
			}

			if param.Schema == nil {
				panic("no schema set in route: " + doc.Pattern)
			}

			if param.Description == "" {
				panic("no description set in route: " + doc.Pattern)
			}
		}
	}

	if doc.OpId == "" {
		panic("no opId set in route: " + doc.Pattern)
	}

	if doc.Pattern == "" {
		panic("no path set in route: " + doc.OpId)
	}

	var schemaName string

	if doc.Resp == nil {
		doc.Resp = DocsSetupData.ErrorStruct
	}

	if doc.RespName != "" {
		schemaName = doc.RespName
	} else {
		schemaName = reflect.TypeOf(doc.Resp).String()
		schemaName = strings.ReplaceAll(schemaName, "docs.", "")
	}

	if schemaName != DocsSetupData.errorStructName {
		if os.Getenv("DEBUG") == "true" {
			fmt.Println(schemaName)
		}

		if _, ok := api.Components.Schemas[schemaName]; !ok {
			schemaRef, err := openapi3gen.NewSchemaRefForValue(doc.Resp, nil, SchemaInject(doc.Resp))

			if err != nil {
				panic(err)
			}

			api.Components.Schemas[schemaName] = schemaRef
		}
	}

	// Add in requests
	var reqBodyRef *Schema
	if doc.Req != nil {
		schemaRef, err := openapi3gen.NewSchemaRefForValue(doc.Req, nil, SchemaInject(doc.Req))

		if err != nil {
			panic(err)
		}

		reqSchemaName := reflect.TypeOf(doc.Req).String()

		if os.Getenv("DEBUG") == "true" {
			fmt.Println("REQUEST:", reqSchemaName)
		}

		api.Components.RequestBodies[doc.Method+"_"+reqSchemaName] = ReqBody{
			Required: true,
			Content: map[string]Content{
				"application/json": {
					Schema: schemaRef,
				},
			},
		}

		if _, ok := api.Paths.Get(doc.Pattern); !ok {
			api.Paths.Set(doc.Pattern, Path{})
		}

		reqBodyRef = &Schema{Ref: "#/components/requestBodies/" + doc.Method + "_" + reqSchemaName}
	}

	operationData := &Operation{
		Tags:        doc.Tags,
		Summary:     doc.Summary,
		Description: doc.Description,
		ID:          doc.OpId,
		Parameters:  doc.Params,
		Responses: map[string]Response{
			"200": {
				Description: "Success",
				Content: map[string]SchemaResp{
					"application/json": {
						Schema: Schema{
							Ref: "#/components/schemas/" + schemaName,
						},
					},
				},
			},
			"400": {
				Description: "Bad Request",
				Content: map[string]SchemaResp{
					"application/json": {
						Schema: Schema{
							Ref: "#/components/schemas/" + DocsSetupData.errorStructName,
						},
					},
				},
			},
		},
	}

	if reqBodyRef != nil {
		operationData.RequestBody = reqBodyRef
	}

	if len(doc.AuthType) == 0 {
		doc.AuthType = []string{}
	}

	operationData.Security = []map[string][]string{}

	for _, auth := range doc.AuthType {
		var authSchema string = auth

		operationData.Security = append(operationData.Security, map[string][]string{
			authSchema: {},
		})
	}

	op, _ := api.Paths.Get(doc.Pattern)

	switch strings.ToLower(doc.Method) {
	case "head":
		op.Head = operationData
	case "get":
		op.Get = operationData
	case "post":
		op.Post = operationData
	case "put":
		op.Put = operationData
	case "patch":
		op.Patch = operationData
	case "delete":
		op.Delete = operationData
	default:
		panic("unknown method: " + doc.Method)
	}

	api.Paths.Set(doc.Pattern, op)
}

func AddWebhook(wdoc *WebhookDoc) {
	schemaRef, err := openapi3gen.NewSchemaRefForValue(wdoc.Format, nil, SchemaInject(wdoc.Format))

	if err != nil {
		panic(err)
	}

	api.Components.RequestBodies[wdoc.FormatName] = ReqBody{
		Required: true,
		Content: map[string]Content{
			"application/json": {
				Schema: schemaRef,
			},
		},
	}

	reqBodyRef := &Schema{Ref: "#/components/requestBodies/" + wdoc.FormatName}

	api.Webhooks.Set(wdoc.Name, Path{
		Post: &Operation{
			ID:          wdoc.Name,
			Tags:        wdoc.Tags,
			Summary:     wdoc.Summary,
			Description: wdoc.Description,
			RequestBody: reqBodyRef,
		},
	})
}

func GetSchema() Openapi {
	return api
}

func SetSchema(new Openapi) {
	api = new
}
