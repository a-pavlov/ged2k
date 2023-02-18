package data

const PIECE_SIZE int = 9728000
const PIECE_SIZE_UINT64 uint64 = 9728000
const BLOCK_SIZE int = 190 * 1024                    // 190kb = PIECE_SIZE/50
const BLOCK_SIZE_UINT64 uint64 = 190 * 1024          // 190kb = PIECE_SIZE/50
const BLOCKS_PER_PIECE int = PIECE_SIZE / BLOCK_SIZE // 50
const HIGHEST_LOWID_ED2K int = 16777216
const REQUEST_QUEUE_SIZE int = 3
const PARTS_IN_REQUEST int = 3

func InBlockOffset(begin uint64, end uint64) (int, int) {
	return int(begin % PIECE_SIZE_UINT64), int(end - begin)
}

func DivCeil64(a uint64, b uint64) uint64 {
	return (a + b - 1) / b
}

func NumPiecesAndBlocks(offset uint64) (int, int) {
	return int(offset / PIECE_SIZE_UINT64), (int)(DivCeil64(offset%PIECE_SIZE_UINT64, BLOCK_SIZE_UINT64))
}
