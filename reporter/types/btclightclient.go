package types

func NewMsgInsertHeaders(
	signer string,
	headers []*IndexedBlock,
) *MsgInsertHeaders {

	headerBytes := make([]BTCHeaderBytes, len(headers))
	for i, h := range headers {
		header := h
		headerBytes[i] = NewBTCHeaderBytesFromBlockHeader(header.Header)
	}

	return &MsgInsertHeaders{
		Signer:  signer,
		Headers: headerBytes,
	}
}
