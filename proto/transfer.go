package proto

import "fmt"

type PieceBlock struct {
	PieceIndex int
	BlockIndex int
}

func (pb *PieceBlock) Get(sb *StateBuffer) *StateBuffer {
	pb.PieceIndex = int(sb.ReadUint32())
	pb.BlockIndex = int(sb.ReadUint32())
	return sb
}

func (pb PieceBlock) Put(sb *StateBuffer) *StateBuffer {
	return sb.Write(uint32(pb.PieceIndex)).Write(uint32(pb.BlockIndex))
}

func (pb PieceBlock) Size() int {
	return DataSize(uint32(pb.PieceIndex)) + DataSize(uint32(pb.BlockIndex))
}

func FromOffset(offset uint64) PieceBlock {
	return PieceBlock{PieceIndex: int(offset / PIECE_SIZE_UINT64), BlockIndex: int(offset%PIECE_SIZE_UINT64) / BLOCK_SIZE}
}

func (pb PieceBlock) Start() uint64 {
	return uint64(pb.PieceIndex)*PIECE_SIZE_UINT64 + uint64(pb.BlockIndex)*uint64(BLOCK_SIZE)
}

func (pb PieceBlock) ToString() string {
	return fmt.Sprintf("[%d:%d]", pb.PieceIndex, pb.BlockIndex)
}

type AddTransferParameters struct {
	Hashes           HashSet
	Filename         ByteContainer
	Filesize         uint64
	Pieces           BitField
	DownloadedBlocks map[int]BitField
}

func (atp *AddTransferParameters) Get(sb *StateBuffer) *StateBuffer {
	sb.Read(&atp.Hashes).Read(&atp.Filename).Read(&atp.Filesize).Read(&atp.Pieces)
	atp.DownloadedBlocks = make(map[int]BitField)
	downloadedBlocksSize := int(sb.ReadUint16())
	if sb.Error() == nil {
		if downloadedBlocksSize > MAX_ELEMS {
			sb.err = fmt.Errorf("downloaded blocks size too large: %v", downloadedBlocksSize)
			return sb
		}

		if downloadedBlocksSize > 0 {
			for i := 0; i < downloadedBlocksSize; i++ {
				pieceIndex := int(sb.ReadUint32())
				bf := BitField{}
				sb.Read(&bf)
				atp.DownloadedBlocks[pieceIndex] = bf
			}
		}
	}

	return sb
}

func (atp AddTransferParameters) Put(sb *StateBuffer) *StateBuffer {
	sb.Write(atp.Hashes).Write(atp.Filename).Write(atp.Filesize).Write(atp.Pieces)
	sb.Write(uint16(len(atp.DownloadedBlocks)))
	for i, x := range atp.DownloadedBlocks {
		sb.Write(uint32(i))
		sb.Write(x)
	}
	return sb
}

func (atp AddTransferParameters) Size() int {
	sz := DataSize(atp.Hashes) +
		DataSize(atp.Filename) +
		DataSize(atp.Filesize) +
		DataSize(atp.Pieces) +
		DataSize(uint16(0))

	for _, x := range atp.DownloadedBlocks {
		sz += DataSize(uint32(0))
		sz += DataSize(x)
	}

	return sz
}

func CreateAddTransferParameters(hash ED2KHash, size uint64, filename string) AddTransferParameters {
	piecesCount, _ := NumPiecesAndBlocks(size)
	return AddTransferParameters{Hashes: HashSet{Hash: hash, PieceHashes: make([]ED2KHash, 0)},
		Filesize:         size,
		Filename:         String2ByteContainer(filename),
		Pieces:           CreateBitField(piecesCount),
		DownloadedBlocks: make(map[int]BitField),
	}
}
