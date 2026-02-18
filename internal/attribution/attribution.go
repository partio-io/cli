package attribution

// Result holds attribution calculation results.
type Result struct {
	TotalLines   int `json:"total_lines"`
	AgentLines   int `json:"agent_lines"`
	HumanLines   int `json:"human_lines"`
	AgentPercent int `json:"agent_percent"`
}
