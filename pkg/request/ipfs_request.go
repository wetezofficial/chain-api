package request

type AddParam struct {
	Recursive         bool   `json:"recursive"  form:"recursive"`
	Progress          bool   `json:"progress"  form:"progress"`
	WrapWithDirectory bool   `json:"wrap-with-directory"  form:"wrap-with-directory"`
	Hidden            bool   `json:"hidden"  form:"hidden"`
	Pin               bool   `json:"pin"  form:"pin"`
	Nocopy            bool   `json:"nocopy"  form:"nocopy"`
	CidVersion        int    `json:"cid-version"  form:"cid-version"`
	Chunker           string `json:"chunker"  form:"chunker"` // string
	WrapDirName       string `json:"-" form:"-"`
	//RawLeaves         bool   `json:"raw-leaves"  form:"raw-leaves"`
	//Fscache           bool   `json:"fscache"  form:"fscache"`
	//Quieter           bool   `json:"quieter"  form:"quieter"`
	//Silent            bool   `json:"silent"  form:"silent"`
	//Quiet             bool   `json:"quiet"  form:"quiet"`
	//Trickle           bool   `json:"trickle"  form:"trickle"`
	//OnlyHash          bool   `json:"only-hash" form:"only-hash"`
	//Hash              string `json:"hash"  form:"hash"` // string
}

//inline [bool]: Inline small blocks into CIDs. (experimental). Required: no.
//inline-limit [int]: Maximum block size to inline. (experimental). Default: 32. Required: no.
//pin [bool]: Pin locally to protect added files from garbage collection. Default: true. Required: no.
//to-files [string]: Add reference to Files API (MFS) at the provided path. Required: no.
