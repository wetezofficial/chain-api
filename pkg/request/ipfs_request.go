package request

type AddParam struct {
	StreamChannels    bool   `json:"stream-channels"  query:"stream-channels"`
	Recursive         bool   `json:"recursive"  query:"recursive"`
	Progress          bool   `json:"progress"  query:"progress"`
	WrapWithDirectory bool   `json:"wrap-with-directory"  query:"wrap-with-directory"`
	Hidden            bool   `json:"hidden"  query:"hidden"`
	Pin               bool   `json:"pin"  query:"pin"`
	Nocopy            bool   `json:"nocopy"  query:"nocopy"`
	CidVersion        int    `json:"cid-version"  query:"cid-version"`
	Chunker           string `json:"chunker"  query:"chunker"` // string
	RawLeaves         bool   `json:"raw-leaves"  query:"raw-leaves"`
	Fscache           bool   `json:"fscache"  query:"fscache"`
	Quieter           bool   `json:"quieter"  query:"quieter"`
	Silent            bool   `json:"silent"  query:"silent"`
	Trickle           bool   `json:"trickle"  query:"trickle"`
	OnlyHash          bool   `json:"only-hash" query:"only-hash"`
	Hash              string `json:"hash"  query:"hash"` // string
	Inline            bool   `json:"inline"  query:"inline"`
	Quiet             bool   `json:"quiet"  query:"quiet"`
	InlineLimit       int    `json:"inline-limit"  query:"inline-limit"`
	ToFiles           string `json:"to-files"  query:"to-files"`
	WrapDirName       string `json:"-"  query:"-"`
}

type PinParam struct {
	Recursive bool   `json:"recursive"  query:"recursive"`
	Progress  bool   `json:"progress"  query:"progress"`
	Arg       string `json:"arg" query:"arg"`
}

type PinLsParam struct {
	Stream bool   `json:"stream"  query:"stream"`
	Quiet  bool   `json:"quiet"  query:"quiet"`
	Type   string `json:"type"  query:"type"` //  The type of pinned keys to list. Can be "direct", "indirect", "recursive", or "all". Default: all. Required: no.
	Arg    string `param:"arg" query:"arg" form:"arg" json:"arg" xml:"arg"`
}

type UpdatePinParam struct {
	Unpin bool     `json:"unpin"  query:"unpin"`
	Arg   []string `json:"arg" query:"arg"`
}
