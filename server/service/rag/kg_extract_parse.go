package rag

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// parseKgExtractJSON 解析图谱抽取 LLM 输出；除标准对象结构外，兼容 LightRAG 式元组关系/实体（如 [["A","B","kw","desc"],...]）。
func parseKgExtractJSON(raw string) (*kgLLMExtractResult, error) {
	s := strings.TrimSpace(raw)
	if m := reJSONFence.FindStringSubmatch(s); len(m) > 1 {
		s = strings.TrimSpace(m[1])
	}
	var wrapper struct {
		PerChunk []json.RawMessage `json:"per_chunk"`
	}
	if err := json.Unmarshal([]byte(s), &wrapper); err != nil {
		return tryLegacyKgExtractParse(s)
	}
	if len(wrapper.PerChunk) == 0 {
		return tryLegacyKgExtractParse(s)
	}
	out := &kgLLMExtractResult{PerChunk: make([]kgLLMChunkExtract, 0, len(wrapper.PerChunk))}
	for _, rawChunk := range wrapper.PerChunk {
		pc, err := parseKgExtractOneChunk(rawChunk)
		if err != nil {
			continue
		}
		out.PerChunk = append(out.PerChunk, pc)
	}
	if len(out.PerChunk) == 0 {
		legacy, lerr := tryLegacyKgExtractParse(s)
		if lerr == nil && len(legacy.PerChunk) > 0 {
			return legacy, nil
		}
		return nil, fmt.Errorf("no valid per_chunk entries")
	}
	return out, nil
}

func tryLegacyKgExtractParse(s string) (*kgLLMExtractResult, error) {
	var legacy kgLLMExtractResult
	if err := json.Unmarshal([]byte(s), &legacy); err != nil {
		return nil, err
	}
	return &legacy, nil
}

func parseKgExtractOneChunk(raw json.RawMessage) (kgLLMChunkExtract, error) {
	var flex struct {
		ChunkIndex    int             `json:"chunk_index"`
		Entities      json.RawMessage `json:"entities"`
		Relationships json.RawMessage `json:"relationships"`
	}
	if err := json.Unmarshal(raw, &flex); err != nil {
		return kgLLMChunkExtract{}, err
	}
	ents := parseKgEntitiesFlexible(flex.Entities)
	rels := parseKgRelationshipsFlexible(flex.Relationships)
	return kgLLMChunkExtract{
		ChunkIndex:    flex.ChunkIndex,
		Entities:      ents,
		Relationships: rels,
	}, nil
}

func parseKgEntitiesFlexible(raw json.RawMessage) []kgLLMEntity {
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 || string(raw) == "null" {
		return nil
	}
	var ents []kgLLMEntity
	if err := json.Unmarshal(raw, &ents); err == nil {
		return ents
	}
	var tuples [][]json.RawMessage
	if err := json.Unmarshal(raw, &tuples); err != nil {
		return nil
	}
	return kgEntitiesFromTupleRows(tuples)
}

func kgEntitiesFromTupleRows(tuples [][]json.RawMessage) []kgLLMEntity {
	var out []kgLLMEntity
	for _, row := range tuples {
		strs := kgRawRowToStrings(row)
		if len(strs) == 0 || strings.TrimSpace(strs[0]) == "" {
			continue
		}
		e := kgLLMEntity{Name: strings.TrimSpace(strs[0])}
		if len(strs) > 1 {
			e.Type = strings.TrimSpace(strs[1])
		}
		if len(strs) > 2 {
			e.Description = strings.TrimSpace(strings.Join(strs[2:], " "))
		}
		out = append(out, e)
	}
	return out
}

func parseKgRelationshipsFlexible(raw json.RawMessage) []kgLLMRelationship {
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 || string(raw) == "null" {
		return nil
	}
	var rels []kgLLMRelationship
	if err := json.Unmarshal(raw, &rels); err == nil {
		return rels
	}
	var tuples [][]json.RawMessage
	if err := json.Unmarshal(raw, &tuples); err == nil {
		return kgRelsFromTupleRows(tuples)
	}
	var strTuples [][]string
	if err := json.Unmarshal(raw, &strTuples); err == nil {
		var rows [][]json.RawMessage
		for _, r := range strTuples {
			row := make([]json.RawMessage, len(r))
			for i, s := range r {
				b, _ := json.Marshal(s)
				row[i] = b
			}
			rows = append(rows, row)
		}
		return kgRelsFromTupleRows(rows)
	}
	return nil
}

func kgRelsFromTupleRows(tuples [][]json.RawMessage) []kgLLMRelationship {
	var out []kgLLMRelationship
	for _, row := range tuples {
		strs := kgRawRowToStrings(row)
		if len(strs) < 2 {
			continue
		}
		r := kgLLMRelationship{
			Source: strings.TrimSpace(strs[0]),
			Target: strings.TrimSpace(strs[1]),
		}
		if len(strs) > 2 {
			r.Keywords = strings.TrimSpace(strs[2])
		}
		if len(strs) > 3 {
			r.Description = strings.TrimSpace(strings.Join(strs[3:], " "))
		}
		out = append(out, r)
	}
	return out
}

func kgRawRowToStrings(row []json.RawMessage) []string {
	var parts []string
	for _, cell := range row {
		var s string
		if json.Unmarshal(cell, &s) == nil {
			parts = append(parts, s)
			continue
		}
		var f float64
		if json.Unmarshal(cell, &f) == nil {
			parts = append(parts, strconv.FormatFloat(f, 'f', -1, 64))
			continue
		}
		parts = append(parts, strings.Trim(string(cell), `"`))
	}
	return parts
}
