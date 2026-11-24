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

func InvokeAs[T any](ctx context.Context, c *Client, name string, in InvokeQueryInput, opts ...ConditionalCallOptions) ([]T, *ReadOutput, error) {
	raw, err := c.Queries().Invoke(ctx, name, in, opts...)
	if err != nil {
		return nil, nil, err
	}
	var res ReadOutput
	if err := json.Unmarshal(raw, &res); err != nil {
		return nil, nil, err
	}
	var out []T
	if err := res.Decode(&out); err != nil {
		return nil, &res, err
	}
	return out, &res, nil
}

func (e ExportRow) Decode(dst any) error {
	rv := reflect.ValueOf(dst)
	if rv.Kind() != reflect.Pointer || rv.IsNil() || rv.Elem().Kind() != reflect.Struct {
		return errors.New("omnigraph: ExportRow.Decode destination must be a non-nil pointer to a struct")
	}
	raw, err := json.Marshal(e.Data)
	if err != nil {
		return err
	}
	var rowMap map[string]json.RawMessage
	if err := json.Unmarshal(raw, &rowMap); err != nil {
		return err
	}
	return decodeStructRow(rowMap, rv.Elem())
}

type TypedStream[T any] struct {
	inner   *Stream[ExportRow]
	current T
	err     error
}

func ExportAs[T any](ctx context.Context, c *Client, in ExportInput, opts ...CallOptions) (*TypedStream[T], error) {
	s, err := c.Export(ctx, in, opts...)
	if err != nil {
		return nil, err
	}
	return &TypedStream[T]{inner: s}, nil
}

func (s *TypedStream[T]) Next() bool {
	if s.err != nil || !s.inner.Next() {
		return false
	}
	var item T
	if err := s.inner.Item().Decode(&item); err != nil {
		s.err = err
		return false
	}
	s.current = item
	return true
}

func (s *TypedStream[T]) Item() T {
	return s.current
}

func (s *TypedStream[T]) Row() ExportRow {
	return s.inner.Item()
}

func (s *TypedStream[T]) Err() error {
	if s.err != nil {
		return s.err
	}
	return s.inner.Err()
}

func (s *TypedStream[T]) Close() error {
	return s.inner.Close()
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
