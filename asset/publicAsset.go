package asset

type PublicAsset struct {
	Asset
	OwnerAssetPool string `json:"ownerAssetPool"`
}

func (pAsset *PublicAsset) Allowance() float64 {
	return pAsset.Value
}
