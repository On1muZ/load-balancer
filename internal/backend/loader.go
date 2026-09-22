package backend

import (
	"encoding/json"
	"fmt"
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
		return nil, fmt.Errorf("invalid path to nodes file")
	}
	var nf nodesFile
	if err := json.Unmarshal(data, &nf); err != nil {
		return nil, fmt.Errorf("invalid json file")
	}
	if len(nf.Nodes) == 0 {
		return nil, fmt.Errorf("no nodes in json file")
	}
	backends := make([]*Backend, 0, len(nf.Nodes))
	for _, n := range nf.Nodes {
		if n.Weight <= 0 {
			return nil, fmt.Errorf("invalid node %s weight %d", n.IP, n.Weight)
		}
		rawURL := "http://" + n.IP
		b, err := NewBackend(rawURL, n.Weight)
		if err != nil {
			return nil, fmt.Errorf("invalid node %w: ", err)
		}
		backends = append(backends, b)
	}
	return backends, nil
}
