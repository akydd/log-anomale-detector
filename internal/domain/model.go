package domain

import "time"

type ClassifiedLogs struct {
	Flag               bool       `json:"flag"`
	Timestamp          *time.Time `json:"timestamp"`
	AnomalyType        *string    `json:"anomaly_type"`
	Severity           *string    `json:"severity"`
	RawEvidence        []string   `json:"raw_evidence"`
	BedrockExplanation *string    `json:"bedrock_explanation"`
}
