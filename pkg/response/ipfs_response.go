package response

type AddResp struct {
	Name           string `json:"Name"`
	Hash           string `json:"Hash"`
	Size           string `json:"Size"`
	WrapWithDirCid string `json:"-"`
	WrapDirName    string `json:"-"`
}

type PinResp struct {
	Cid  string `json:"Cid"`
	Type string `json:"Type"`
}

type PinListMapResult struct {
	Keys map[string]PinResp `json:"Keys"`
}

type IpfsObjectStat struct {
	Hash           string `json:"Hash"`
	NumLinks       int    `json:"NumLinks"`
	BlockSize      int    `json:"BlockSize"`
	LinksSize      int    `json:"LinksSize"`
	DataSize       int    `json:"DataSize"`
	CumulativeSize int    `json:"CumulativeSize"`
}
