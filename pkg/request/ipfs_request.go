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
}

type PinParam struct {
	Recursive bool   `json:"recursive"  form:"recursive"`
	Progress  bool   `json:"progress"  form:"progress"`
	Arg       string `json:"arg"  form:"arg"`
}
