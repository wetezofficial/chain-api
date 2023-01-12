package request

type AddParam struct {
	StreamChannels    bool   `json:"stream-channels"  form:"stream-channels"`
	Recursive         bool   `json:"recursive"  form:"recursive"`
	Progress          bool   `json:"progress"  form:"progress"`
	WrapWithDirectory bool   `json:"wrap-with-directory"  form:"wrap-with-directory"`
	Hidden            bool   `json:"hidden"  form:"hidden"`
	Pin               bool   `json:"pin"  form:"pin"`
	Nocopy            bool   `json:"nocopy"  form:"nocopy"`
	CidVersion        int    `json:"cid-version"  form:"cid-version"`
	Chunker           string `json:"chunker"  form:"chunker"` // string
	RawLeaves         bool   `json:"raw-leaves"  form:"raw-leaves"`
	Fscache           bool   `json:"fscache"  form:"fscache"`
	Quieter           bool   `json:"quieter"  form:"quieter"`
	Silent            bool   `json:"silent"  form:"silent"`
	Trickle           bool   `json:"trickle"  form:"trickle"`
	OnlyHash          bool   `json:"only-hash" form:"only-hash"`
	Hash              string `json:"hash"  form:"hash"` // string
	Inline            bool   `json:"inline"  form:"inline"`
	Quiet             bool   `json:"quiet"  form:"quiet"`
	InlineLimit       int    `json:"inline-limit"  form:"inline-limit"`
	ToFiles           string `json:"to-files"  form:"to-files"`
	WrapDirName       string `json:"-"  form:"-"`
}

type PinParam struct {
	Recursive bool `json:"recursive"  form:"recursive"`
}
