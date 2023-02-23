package main

import (
	"fmt"
	"github.com/a-pavlov/ged2k/proto"
	"hash"
)

type ReceivingPiece struct {
	hash           hash.Hash
	blocks         []*PendingBlock
	hashBlockIndex int
}

func (rp *ReceivingPiece) InsertBlock(pb *PendingBlock) {
	skipBlocks := 0
	for _, x := range rp.blocks {
		if x.block.BlockIndex < pb.block.BlockIndex {
			skipBlocks++
		} else {
			break
		}
	}

	fmt.Println("skip blocks", skipBlocks)

	switch skipBlocks {
	case 0:
		rp.blocks = append([]*PendingBlock{pb}, rp.blocks...)
	case len(rp.blocks):
		rp.blocks = append(rp.blocks, pb)
	default:
		rp.blocks = append(rp.blocks[:skipBlocks+1], rp.blocks[skipBlocks:]...)
		rp.blocks[skipBlocks] = pb
	}

	for _, x := range rp.blocks {
		// skip blocks with index less than start hashing
		if x.block.BlockIndex < rp.hashBlockIndex {
			continue
		}

		if rp.hashBlockIndex != x.block.BlockIndex {
			break
		}

		rp.hash.Write(x.data)
		rp.hashBlockIndex++
	}
}

func (rp *ReceivingPiece) Hash() proto.EMuleHash {
	h := proto.EMuleHash{}
	rp.hash.Sum(h[:])
	return h
}
