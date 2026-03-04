package sdk

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"okpay/payment/plugin/contract"
)

// DecodeLosslessJSON decodes JSON into a lossless Value.
// Numbers are never converted to float64.
func DecodeLosslessJSON(raw []byte) (contract.Value, error) {
	if len(bytes.TrimSpace(raw)) == 0 {
		return contract.NullValue(), nil
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	var v any
	if err := dec.Decode(&v); err != nil {
		return contract.Value{}, fmt.Errorf("decode lossless json failed: %w", err)
	}
	return AnyToValue(v)
}

// DecodeLosslessJSONObject decodes a JSON object into ObjectValue.
func DecodeLosslessJSONObject(raw []byte) (*contract.ObjectValue, error) {
	v, err := DecodeLosslessJSON(raw)
	if err != nil {
		return nil, err
	}
	if v.Kind != contract.ValueKindObject || v.Object == nil {
		return nil, fmt.Errorf("json root is not object")
	}
	return v.Object, nil
}

// AnyToValue converts plain Go values into lossless Value.
// It avoids json round-trip when possible.
func AnyToValue(in any) (contract.Value, error) {
	if in == nil {
		return contract.NullValue(), nil
	}
	switch v := in.(type) {
	case contract.Value:
		return v, nil
	case *contract.Value:
		if v == nil {
			return contract.NullValue(), nil
		}
		return *v, nil
	case time.Time:
		return contract.StringValue(v.Format(time.RFC3339Nano)), nil
	case *time.Time:
		if v == nil {
			return contract.NullValue(), nil
		}
		return contract.StringValue(v.Format(time.RFC3339Nano)), nil
	case json.Number:
		return parseLosslessNumber(v)
	}
	return anyToValueReflect(reflect.ValueOf(in))
}

func rawMessageToValue(raw json.RawMessage) (contract.Value, error) {
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 {
		return contract.NullValue(), nil
	}
	return DecodeLosslessJSON(trimmed)
}

func anyToValueReflect(rv reflect.Value) (contract.Value, error) {
	if !rv.IsValid() {
		return contract.NullValue(), nil
	}
	for rv.Kind() == reflect.Interface || rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			return contract.NullValue(), nil
		}
		rv = rv.Elem()
	}
	// Keep JSON semantic for json.RawMessage even when it appears as a struct field.
	if rv.Type() == reflect.TypeOf(json.RawMessage{}) {
		raw, _ := rv.Interface().(json.RawMessage)
		return rawMessageToValue(raw)
	}

	switch rv.Kind() {
	case reflect.String:
		return contract.StringValue(rv.String()), nil
	case reflect.Bool:
		return contract.BoolValue(rv.Bool()), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return contract.Int64Value(rv.Int()), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return contract.UInt64Value(rv.Uint()), nil
	case reflect.Float32:
		return contract.DecimalValue(strconv.FormatFloat(rv.Float(), 'f', -1, 32)), nil
	case reflect.Float64:
		return contract.DecimalValue(strconv.FormatFloat(rv.Float(), 'f', -1, 64)), nil
	case reflect.Slice:
		if rv.IsNil() {
			return contract.NullValue(), nil
		}
		// Preserve raw bytes for binary payloads.
		if rv.Type().Elem().Kind() == reflect.Uint8 {
			out := make([]byte, rv.Len())
			reflect.Copy(reflect.ValueOf(out), rv)
			return contract.BytesValue(out), nil
		}
		return arrayToLosslessValue(rv)
	case reflect.Array:
		return arrayToLosslessValue(rv)
	case reflect.Map:
		return mapToLosslessValue(rv)
	case reflect.Struct:
		// time.Time should already be handled by AnyToValue fast-path.
		return structToLosslessValue(rv)
	default:
		return contract.Value{}, fmt.Errorf("unsupported value type: %s", rv.Type())
	}
}

func arrayToLosslessValue(rv reflect.Value) (contract.Value, error) {
	out := make([]contract.Value, 0, rv.Len())
	for i := 0; i < rv.Len(); i++ {
		val, err := anyToValueReflect(rv.Index(i))
		if err != nil {
			return contract.Value{}, err
		}
		out = append(out, val)
	}
	return contract.ArrayValue(out), nil
}

func mapToLosslessValue(rv reflect.Value) (contract.Value, error) {
	if rv.IsNil() {
		return contract.NullValue(), nil
	}
	fields := make(map[string]contract.Value, rv.Len())
	iter := rv.MapRange()
	for iter.Next() {
		key := iter.Key()
		if key.Kind() != reflect.String {
			return contract.Value{}, fmt.Errorf("unsupported map key type: %s", key.Type())
		}
		val, err := anyToValueReflect(iter.Value())
		if err != nil {
			return contract.Value{}, fmt.Errorf("decode key %s failed: %w", key.String(), err)
		}
		fields[key.String()] = val
	}
	return contract.ObjectMapValue(fields), nil
}

func structToLosslessValue(rv reflect.Value) (contract.Value, error) {
	rt := rv.Type()
	fields := make(map[string]contract.Value, rt.NumField())
	for i := 0; i < rt.NumField(); i++ {
		sf := rt.Field(i)
		if sf.PkgPath != "" { // unexported
			continue
		}
		name, omitEmpty, skip := parseJSONFieldTag(sf)
		if skip {
			continue
		}
		fv := rv.Field(i)
		if omitEmpty && isEmptyValue(fv) {
			continue
		}
		val, err := anyToValueReflect(fv)
		if err != nil {
			return contract.Value{}, fmt.Errorf("decode field %s failed: %w", sf.Name, err)
		}
		fields[name] = val
	}
	return contract.ObjectMapValue(fields), nil
}

func parseJSONFieldTag(sf reflect.StructField) (name string, omitEmpty bool, skip bool) {
	tag := sf.Tag.Get("json")
	if tag == "-" {
		return "", false, true
	}
	if tag == "" {
		return sf.Name, false, false
	}
	parts := strings.Split(tag, ",")
	fieldName := strings.TrimSpace(parts[0])
	if fieldName == "" {
		fieldName = sf.Name
	}
	for _, p := range parts[1:] {
		if strings.TrimSpace(p) == "omitempty" {
			omitEmpty = true
		}
	}
	return fieldName, omitEmpty, false
}

func isEmptyValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	case reflect.Struct:
		return v.IsZero()
	default:
		return false
	}
}

func parseLosslessNumber(n json.Number) (contract.Value, error) {
	s := strings.TrimSpace(n.String())
	if s == "" {
		return contract.Value{}, fmt.Errorf("empty number")
	}
	// Keep non-integer numbers as exact decimal string.
	if strings.ContainsAny(s, ".eE") {
		return contract.DecimalValue(s), nil
	}
	if i, err := strconv.ParseInt(s, 10, 64); err == nil {
		return contract.Int64Value(i), nil
	}
	if u, err := strconv.ParseUint(s, 10, 64); err == nil {
		return contract.UInt64Value(u), nil
	}
	return contract.Value{}, fmt.Errorf("integer out of range: %s", s)
}

// ValueToAny converts lossless Value into plain Go values.
// Decimal is preserved as string to avoid float precision loss.
func ValueToAny(v contract.Value) (any, error) {
	switch v.Kind {
	case contract.ValueKindNull:
		return nil, nil
	case contract.ValueKindString:
		return v.String, nil
	case contract.ValueKindBool:
		return v.Bool, nil
	case contract.ValueKindInt64:
		return v.Int64, nil
	case contract.ValueKindUInt64:
		return v.UInt64, nil
	case contract.ValueKindDecimal:
		return v.Decimal, nil
	case contract.ValueKindBytes:
		out := make([]byte, len(v.Bytes))
		copy(out, v.Bytes)
		return out, nil
	case contract.ValueKindObject:
		if v.Object == nil || len(v.Object.Fields) == 0 {
			return map[string]any{}, nil
		}
		out := make(map[string]any, len(v.Object.Fields))
		for k, child := range v.Object.Fields {
			val, err := ValueToAny(child)
			if err != nil {
				return nil, err
			}
			out[k] = val
		}
		return out, nil
	case contract.ValueKindArray:
		out := make([]any, 0, len(v.Array))
		for _, item := range v.Array {
			val, err := ValueToAny(item)
			if err != nil {
				return nil, err
			}
			out = append(out, val)
		}
		return out, nil
	default:
		return nil, fmt.Errorf("unsupported value kind: %s", v.Kind)
	}
}

// ValueMapToAnyMap converts map[string]Value into map[string]any.
func ValueMapToAnyMap(in map[string]contract.Value) (map[string]any, error) {
	if len(in) == 0 {
		return map[string]any{}, nil
	}
	out := make(map[string]any, len(in))
	for k, v := range in {
		val, err := ValueToAny(v)
		if err != nil {
			return nil, err
		}
		out[k] = val
	}
	return out, nil
}
