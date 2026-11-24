package omnigraph

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
)

type Params = map[string]any

func QueryAs[T any](ctx context.Context, c *Client, in QueryInput, opts ...CallOptions) ([]T, *ReadOutput, error) {
	res, err := c.Query(ctx, in, opts...)
	if err != nil {
		return nil, nil, err
	}
	var out []T
	if err := res.Decode(&out); err != nil {
		return nil, res, err
	}
	return out, res, nil
}

func (r *ReadOutput) Decode(dst any) error {
	if r == nil || r.Rows == nil {
		return nil
	}
	rv := reflect.ValueOf(dst)
	if rv.Kind() != reflect.Pointer || rv.IsNil() || rv.Elem().Kind() != reflect.Slice {
		return errors.New("omnigraph: Decode destination must be a non-nil pointer to a slice")
	}

	raw, err := json.Marshal(r.Rows)
	if err != nil {
		return err
	}

	var rawRows []map[string]json.RawMessage
	if err := json.Unmarshal(raw, &rawRows); err != nil {
		return err
	}

	sliceVal := rv.Elem()
	elemType := sliceVal.Type().Elem()
	isPtr := elemType.Kind() == reflect.Pointer
	structType := elemType
	if isPtr {
		structType = elemType.Elem()
	}

	if structType.Kind() != reflect.Struct {
		return json.Unmarshal(raw, dst)
	}

	out := reflect.MakeSlice(sliceVal.Type(), len(rawRows), len(rawRows))
	for i, rowMap := range rawRows {
		elemPtr := reflect.New(structType)
		if err := decodeStructRow(rowMap, elemPtr.Elem()); err != nil {
			return fmt.Errorf("omnigraph: row %d: %w", i, err)
		}
		if isPtr {
			out.Index(i).Set(elemPtr)
		} else {
			out.Index(i).Set(elemPtr.Elem())
		}
	}
	sliceVal.Set(out)
	return nil
}

func decodeStructRow(rowMap map[string]json.RawMessage, dst reflect.Value) error {
	t := dst.Type()
	for i := 0; i < t.NumField(); i++ {
		sf := t.Field(i)
		if !sf.IsExported() {
			continue
		}
		col := columnNameForField(sf)
		if col == "-" {
			continue
		}
		rawVal, ok := rowMap[col]
		if !ok {
			continue
		}
		if err := json.Unmarshal(rawVal, dst.Field(i).Addr().Interface()); err != nil {
			return fmt.Errorf("field %s (%s): %w", sf.Name, col, err)
		}
	}
	return nil
}

func columnNameForField(sf reflect.StructField) string {
	if tag, ok := sf.Tag.Lookup("omnigraph"); ok {
		name := strings.Split(tag, ",")[0]
		if name != "" {
			return name
		}
	}
	if tag, ok := sf.Tag.Lookup("json"); ok {
		name := strings.Split(tag, ",")[0]
		if name != "" {
			return name
		}
	}
	return sf.Name
}
