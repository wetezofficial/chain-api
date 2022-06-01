package openrpc

import (
	"fmt"
	"github.com/go-openapi/spec"
	"path"
	"reflect"
	"starnet/chain-api/pkg/utils"
)

const (
	params = "Params"
	result = "Result"
)

func persistTitleAndDesc(prev, next spec.Schema) spec.Schema {
	next.Title = utils.FirstOf(next.Title, prev.Title)
	next.Description = utils.FirstOf(next.Description, prev.Description)
	return next
}

func persistFields(prev, next spec.Schema) spec.Schema {
	next.Title = utils.FirstOf(next.Title, prev.Title, path.Base(prev.Ref.String()))
	next.Description = utils.FirstOf(next.Description, prev.Description)
	if next.Items == nil {
		next.Items = prev.Items
	}
	return next
}

func resolveSchema(openrpc *OpenRPCSpec1, sch spec.Schema) spec.Schema {
	doc, _, _ := sch.Ref.GetPointer().Get(openrpc)

	if s, ok := doc.(spec.Schema); ok {
		sch = persistFields(sch, s)
	} else if cd, ok := doc.(*ContentDescriptor); ok {
		sch = persistFields(sch, cd.Schema)
	}

	if sch.Ref.GetURL() != nil {
		return resolveSchema(openrpc, sch)
	}
	return sch
}

func getConcreteType(in string) string {
	switch in {
	case reflect.Bool.String(), "boolean":
		return reflect.Bool.String()
	default:
		return in
	}
}

func getObjectType(openrpc *OpenRPCSpec1, sch spec.Schema) string {
	sch = resolveSchema(openrpc, sch)

	if len(sch.Properties) > 0 || len(sch.Type) < 1 {
		return utils.CamelCase(sch.Title)
	}

	return getConcreteType(sch.Type[0])
}

func dereference(openrpc *OpenRPCSpec1, name string, sch spec.Schema, om *ObjectMap) {
	// resolve all pointers
	sch = resolveSchema(openrpc, sch)

	if len(sch.Properties) > 0 {
		for key, value := range sch.Properties {
			value.Title = key
			dereference(openrpc, sch.Title, value, om)
		}
		om.Set(name, BasicType{sch.Description, sch.Title, utils.CamelCase(sch.Title)})
		return
	} else if len(sch.OneOf) > 0 {
		next := sch.OneOf[0]
		dereference(openrpc, sch.Title, next, om)
		om.Set(name, BasicType{sch.Description, sch.Title, getObjectType(openrpc, resolveSchema(openrpc, next))})
		return
	} else if sch.Items != nil {
		if sch.Items.Schema != nil {
			dereference(openrpc, sch.Title, *sch.Items.Schema, om)
			dereference(openrpc, name, persistTitleAndDesc(sch, *sch.Items.Schema), om)
			om.Set(name, BasicType{sch.Description, sch.Title, fmt.Sprintf("[]%s", getObjectType(openrpc, persistTitleAndDesc(sch, *sch.Items.Schema)))})
		} else if len(sch.Items.Schemas) > 0 {
			om.Set(name, BasicType{sch.Description, sch.Title, "[]string"})
		}
		return
	}

	if len(sch.Type) == 0 {
		return
	}

	om.Set(name, BasicType{sch.Description, sch.Title, getConcreteType(sch.Type[0])})
	return
}

// GetTypes constructs all possible type definitions from the spec
func GetTypes(openrpc *OpenRPCSpec1, om *ObjectMap) {
	for _, m := range openrpc.Methods {
		name := fmt.Sprintf("%s%s", utils.CamelCase(m.Name), params)
		for _, param := range m.Params {
			sch := param.Schema
			sch.Title = utils.FirstOf(sch.Title, param.Name)
			sch.Description = utils.FirstOf(sch.Description, param.Description)
			dereference(openrpc, name, sch, om)
		}
		if m.Result != nil {
			name = fmt.Sprintf("%s%s", utils.CamelCase(m.Name), result)
			res := m.Result
			sch := res.Schema
			sch.Title = utils.FirstOf(sch.Title, res.Name)
			sch.Description = utils.FirstOf(sch.Description, res.Description)
			dereference(openrpc, name, sch, om)
		}
	}
}
