package response

type AddResp struct {
	Name string `json:"Name"`
	Hash string `json:"Hash"`
	Size string `json:"Size"`
}

type PinResp struct {
	Cid  string `json:"Cid"`
	Type string `json:"Type"`
}

type PinListMapResult struct {
	Keys map[string]PinResp `json:"Keys"`
}
