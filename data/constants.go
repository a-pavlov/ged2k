package data

import "github.com/a-pavlov/ged2k/proto"

func InBlockOffset(begin uint64) int {
	return int(begin % proto.PIECE_SIZE_UINT64)
}

func DivCeil64(a uint64, b uint64) uint64 {
	return (a + b - 1) / b
}

func NumPiecesAndBlocks(offset uint64) (int, int) {
	return int(offset / proto.PIECE_SIZE_UINT64), (int)(DivCeil64(offset%proto.PIECE_SIZE_UINT64, proto.BLOCK_SIZE_UINT64))
}
