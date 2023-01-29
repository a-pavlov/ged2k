package data

const PIECE_SIZE int = 9728000
const BLOCK_SIZE int = 190 * 1024                    // 190kb = PIECE_SIZE/50
const BLOCKS_PER_PIECE int = PIECE_SIZE / BLOCK_SIZE // 50
const HIGHEST_LOWID_ED2K int = 16777216
const REQUEST_QUEUE_SIZE int = 3
const PARTS_IN_REQUEST int = 3
