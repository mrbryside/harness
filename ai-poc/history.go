package main

import (
	"encoding/json"
	"sync"
	"time"
)

type APIRound struct {
	ID        string      `json:"id"`
	Timestamp time.Time   `json:"timestamp"`
	Request   *APIRequest  `json:"request"`
	Response  *APIResponse `json:"response"`
}

type APIRequest struct {
	Model    string                 `json:"model"`
	Messages []json.RawMessage      `json:"messages"`
	Tools    []json.RawMessage      `json:"tools,omitempty"`
	Stream   bool                   `json:"stream"`
}

type APIResponse struct {
	Text           string              `json:"text,omitempty"`
	Reasoning      string              `json:"reasoning,omitempty"`
	ToolCalls      []ToolCallInfo      `json:"tool_calls,omitempty"`
	RawChunks      []map[string]interface{} `json:"raw_chunks"`
	FinishReason   string              `json:"finish_reason,omitempty"`
}

type ToolCallInfo struct {
	ID        string `json:"id"`
	Type      string `json:"type"`
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
	Result    string `json:"result,omitempty"`
}

type HistoryStore struct {
	mu     sync.RWMutex
	rounds []APIRound
}

func NewHistoryStore() *HistoryStore {
	return &HistoryStore{
		rounds: make([]APIRound, 0),
	}
}

func (h *HistoryStore) AddRound(round APIRound) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.rounds = append(h.rounds, round)
}

func (h *HistoryStore) GetAll() []APIRound {
	h.mu.RLock()
	defer h.mu.RUnlock()
	result := make([]APIRound, len(h.rounds))
	copy(result, h.rounds)
	return result
}

func (h *HistoryStore) Clear() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.rounds = make([]APIRound, 0)
}

func (h *HistoryStore) GetLast() *APIRound {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if len(h.rounds) == 0 {
		return nil
	}
	last := h.rounds[len(h.rounds)-1]
	return &last
}
