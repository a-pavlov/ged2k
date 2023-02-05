package data

const PIECE_SIZE int = 9728000
const PIECE_SIZE_UINT64 uint64 = 9728000
const BLOCK_SIZE int = 190 * 1024                    // 190kb = PIECE_SIZE/50
const BLOCKS_PER_PIECE int = PIECE_SIZE / BLOCK_SIZE // 50
const HIGHEST_LOWID_ED2K int = 16777216
const REQUEST_QUEUE_SIZE int = 3
const PARTS_IN_REQUEST int = 3

func BeginEnd2StartLength(begin uint64, end uint64) (int, uint64, uint64) {
	pieceIndex := int(begin / PIECE_SIZE_UINT64)
	start := begin % PIECE_SIZE_UINT64
	length := end - begin
	if length > PIECE_SIZE_UINT64 {
		panic("chank size greater than PIECE_SIZE")
	}

	return pieceIndex, start, length
}

func Offset2PieceBlock(offset uint64) (int, int) {
	piece := (int)(offset / PIECE_SIZE_UINT64)
	start := (int)(offset % PIECE_SIZE_UINT64)
	return piece, start / BLOCK_SIZE
	//return new PieceBlock(piece, (int)(start / Constants.BLOCK_SIZE));
}
