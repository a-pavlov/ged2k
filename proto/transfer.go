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

type AddTransferParameters struct {
	Hashes           HashSet
	Filename         ByteContainer
	Filesize         uint64
	Pieces           BitField
	DownloadedBlocks []PieceBlock
}

func (atp *AddTransferParameters) Get(sb *StateBuffer) *StateBuffer {
	sb.Read(&atp.Hashes).Read(&atp.Filename).Read(&atp.Filesize).Read(&atp.Pieces)
	downloadedBlocksSize := sb.ReadUint16()
	if sb.Error() == nil {
		if int(downloadedBlocksSize) > MAX_ELEMS {
			sb.err = fmt.Errorf("downloaded blocks size too large: %v", downloadedBlocksSize)
			return sb
		}

		atp.DownloadedBlocks = make([]PieceBlock, int(downloadedBlocksSize))
		for i, _ := range atp.DownloadedBlocks {
			sb.Read(&atp.DownloadedBlocks[i])
		}
	}

	return sb
}

func (atp AddTransferParameters) Put(sb *StateBuffer) *StateBuffer {
	sb.Write(atp.Hashes).Write(atp.Filename).Write(atp.Filesize).Write(atp.Pieces)
	sb.Write(uint16(len(atp.DownloadedBlocks)))
	for _, x := range atp.DownloadedBlocks {
		sb.Write(x)
	}
	return sb
}

func (atp AddTransferParameters) Size() int {
	return DataSize(atp.Hashes) +
		DataSize(atp.Filename) +
		DataSize(atp.Filesize) +
		DataSize(atp.Pieces) +
		DataSize(uint16(0)) +
		len(atp.DownloadedBlocks)*DataSize(PieceBlock{})
}
