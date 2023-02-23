package proto

type TransferResumeData struct {
	Hashes HashSet
	Pieces BitField
}

type AddTransferParameters struct {
}
