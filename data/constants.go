package data

import "github.com/a-pavlov/ged2k/proto"

func InBlockOffset(begin uint64) int {
	return int(begin % proto.PIECE_SIZE_UINT64)
}

func DivCeil64(a uint64, b uint64) uint64 {
	return (a + b - 1) / b
}

func NumPiecesAndBlocks(offset uint64) (int, int) {
	if offset == 0 {
		return 0, 0
	}
	blocksInLastPiece := (int)(DivCeil64(offset%proto.PIECE_SIZE_UINT64, proto.BLOCK_SIZE_UINT64))
	if blocksInLastPiece == 0 {
		blocksInLastPiece = proto.BLOCKS_PER_PIECE
	}
	return int(DivCeil64(offset, proto.PIECE_SIZE_UINT64)), blocksInLastPiece
}
