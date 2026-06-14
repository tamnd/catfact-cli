package catfact

// Fact is a single cat fact entry.
type Fact struct {
	Rank   int    `json:"rank"`
	Fact   string `json:"fact" kit:"id"`
	Length int    `json:"length"`
}

// unexported: only used inside catfact.go for JSON decode

type factResponse struct {
	Fact   string `json:"fact"`
	Length int    `json:"length"`
}

type factsPage struct {
	Data []factResponse `json:"data"`
}
