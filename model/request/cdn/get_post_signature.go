package cdn

type GetPostSignatureRequest struct {
	SubDir string `form:"sub_dir"`
}
type GetPostSignatureResponse struct {
	Policy           string `json:"policy"`
	SecurityToken    string `json:"x-oss-security-token"`
	SignatureVersion string `json:"x_oss_signature_version"`
	Credential       string `json:"x_oss_credential"`
	Date             string `json:"x_oss_date"`
	Signature        string `json:"x-oss-signature"`
	Host             string `json:"host"`
	Dir              string `json:"dir"`
}
