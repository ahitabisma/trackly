package investor

type DuplicateCandidate struct {
	InvestorAID   uint    `json:"investor_a_id"`
	InvestorAName string  `json:"investor_a_name"`
	InvestorBID   uint    `json:"investor_b_id"`
	InvestorBName string  `json:"investor_b_name"`
	Score         float64 `json:"score"`
}
