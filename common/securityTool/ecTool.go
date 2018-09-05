package securityTool

type ECTool struct {
}

func (ec *ECTool) ParsePublicKey(publicKey string) (interface{}, error) {
	return nil, nil
}
func (ec *ECTool) EncryptByPoolPublicKey(publicKey []byte, data []byte) (string, error) {
	return "", nil
}
func (ec *ECTool) VerifySignStrByPoolPublicKey(data []byte, signature, publicKey string) (bool, error) {
	return false, nil
}
func (ec *ECTool) VerifyTxByPoolPublicKey(publicKey string, obj interface{}) (bool, error) {
	return false, nil
}
