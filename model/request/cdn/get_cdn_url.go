package cdn

type GetCdnUrlRequest struct {
	Url string `form:"url" binding:"required"`
}

type GetCdnUrlResponse struct {
	Url string `json:"url" binding:"required"`
}
