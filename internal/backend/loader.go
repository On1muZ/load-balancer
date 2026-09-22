package backend

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
)

type nodeEntry struct {
	IP     string `json:"ip"`
	Weight int    `json:"weight"`
}

type nodesFile struct {
	Nodes []nodeEntry `json:"nodes"`
}

func LoadNodesFromFile(path string) ([]*Backend, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read nodes file: %w", err)
	}
	var nf nodesFile
	if err := decodeStrict(data, &nf); err != nil {
		return nil, fmt.Errorf("invalid json: %w", err)
	}
	if len(nf.Nodes) == 0 {
		return nil, fmt.Errorf("no nodes in json file")
	}
	backends := make([]*Backend, 0, len(nf.Nodes))
	seen := make(map[string]struct{}, len(nf.Nodes))
	for i, n := range nf.Nodes {
		if n.IP == "" {
			return nil, fmt.Errorf("node #%d: empty ip", i+1)
		}
		if _, ok := seen[n.IP]; ok {
			return nil, fmt.Errorf("node #%d: duplicate ip %s", i+1, n.IP)
		}
		seen[n.IP] = struct{}{}
		if n.Weight <= 0 || n.Weight > math.MaxInt32 {
			return nil, fmt.Errorf("node #%d (%s): invalid weight %d, must be in 1..%d", i+1, n.IP, n.Weight, math.MaxInt32)
		}
		b, err := NewBackend("http://"+n.IP, n.Weight)
		if err != nil {
			return nil, fmt.Errorf("node #%d (%s): %w", i+1, n.IP, err)
		}
		backends = append(backends, b)
	}
	return backends, nil
}

func decodeStrict(data []byte, v any) error {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		var syntaxErr *json.SyntaxError
		var typeErr *json.UnmarshalTypeError
		switch {
		case errors.As(err, &syntaxErr):
			line, col := position(data, syntaxErr.Offset)
			return fmt.Errorf("line %d, column %d: %w", line, col, err)
		case errors.As(err, &typeErr):
			line, col := position(data, typeErr.Offset)
			return fmt.Errorf("line %d, column %d: field %q must be %s, got %s", line, col, typeErr.Field, typeErr.Type, typeErr.Value)
		case errors.Is(err, io.ErrUnexpectedEOF), errors.Is(err, io.EOF):
			return fmt.Errorf("unexpected end of file")
		}
		return err
	}
	if _, err := dec.Token(); err != io.EOF {
		line, col := position(data, dec.InputOffset())
		return fmt.Errorf("line %d, column %d: unexpected data after json object", line, col)
	}
	return nil
}

func position(data []byte, offset int64) (line, col int) {
	if offset > int64(len(data)) {
		offset = int64(len(data))
	}
	before := data[:offset]
	line = bytes.Count(before, []byte("\n")) + 1
	col = int(offset) - bytes.LastIndexByte(before, '\n')
	return line, col
}
