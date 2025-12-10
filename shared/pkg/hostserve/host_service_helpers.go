package hostserve

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"
)

func newUUID() uuid.UUID {
	id, err := uuid.NewV7()
	if err != nil {
		id = uuid.New()
	}
	return id
}

func extractRequestFields(prefix string, req proto.Message, depth int) []interface{} {
	if req == nil || depth > 10 {
		return nil
	}

	fields := []interface{}{}
	v := reflect.ValueOf(req)
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil
		}
		v = v.Elem()
	}

	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		value := v.Field(i)

		// Get the protobuf field name from protobuf tag, fallback to json tag
		fieldName := getProtoFieldName(field)
		if fieldName == "" || fieldName == "-" {
			continue
		}

		// Build the full field path
		fullName := fieldName
		if prefix != "" {
			fullName = prefix + "." + fieldName
		}

		// Extract the value appropriately
		logValue, recurse := extractFieldValue(value)

		if recurse != nil {
			// Recursively handle nested message
			nestedFields := extractRequestFields(fullName, recurse, depth+1)
			fields = append(fields, nestedFields...)
		} else if logValue != nil {
			fields = append(fields, fullName, logValue)
		}
	}

	return fields
}

func getProtoFieldName(field reflect.StructField) string {
	// Try protobuf tag first
	if protoTag := field.Tag.Get("protobuf"); protoTag != "" {
		for _, part := range strings.Split(protoTag, ",") {
			if strings.HasPrefix(part, "name=") {
				return strings.TrimPrefix(part, "name=")
			}
		}
	}

	// Fall back to json tag
	if jsonTag := field.Tag.Get("json"); jsonTag != "" {
		return strings.Split(jsonTag, ",")[0]
	}

	return ""
}

func extractFieldValue(value reflect.Value) (interface{}, proto.Message) {
	switch value.Kind() {
	case reflect.String:
		s := value.String()
		if s == "" {
			return nil, nil
		}
		return s, nil

	case reflect.Int, reflect.Int32, reflect.Int64:
		i := value.Int()
		if i == 0 {
			return nil, nil
		}
		return i, nil

	case reflect.Uint, reflect.Uint32, reflect.Uint64:
		u := value.Uint()
		if u == 0 {
			return nil, nil
		}
		return u, nil

	case reflect.Bool:
		b := value.Bool()
		if !b {
			return nil, nil
		}
		return b, nil

	case reflect.Slice:
		if value.IsNil() || value.Len() == 0 {
			return nil, nil
		}

		if value.Type().Elem().Kind() == reflect.Uint8 {
			// Byte slice (data) - log length not content
			return fmt.Sprintf("<%d bytes>", value.Len()), nil
		}

		// For other slices, we could iterate and recurse, but for now just log count
		return fmt.Sprintf("<%d items>", value.Len()), nil

	case reflect.Ptr:
		if value.IsNil() {
			return nil, nil
		}

		// Check if it implements proto.Message (nested message)
		if msg, ok := value.Interface().(proto.Message); ok {
			return nil, msg
		}

		// Otherwise dereference and try again
		return extractFieldValue(value.Elem())

	case reflect.Struct:
		// Check if it implements proto.Message
		if value.CanAddr() {
			if msg, ok := value.Addr().Interface().(proto.Message); ok {
				return nil, msg
			}
		}
		return fmt.Sprintf("<%s>", value.Type().Name()), nil

	default:
		return fmt.Sprintf("<%s>", value.Kind()), nil
	}
}
