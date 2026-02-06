package response

type UploadInstructionResponse struct {
    UploadURL string `json:"uploadUrl"`
    PublicURL string `json:"publicUrl"`
    Exists    bool   `json:"exists"`
}