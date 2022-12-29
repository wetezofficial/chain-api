package response

type AddResp struct {
	Name string `json:"Name"`
	Hash string `json:"Hash"`
	Size uint64 `json:"Size"`
}
