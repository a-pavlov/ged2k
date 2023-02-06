package data

type Range struct {
	Begin uint64
	End   uint64
}

type Region struct {
	Segments []Range
}

func Make(begin uint64, end uint64) Range {
	return Range{Begin: begin, End: end}
}

func MakeRegion(r Range) Region {
	return Region{Segments: []Range{r}}
}

func (region *Region) ShrinkEnd(size uint64) {
	if len(region.Segments) == 0 {
		panic("no Segments to shrink")
	}

	region.Segments[0].End = region.Segments[0].Begin + size
}

func (region *Region) Sub(seg Range) {
	res := []Range{}
	for _, x := range region.Segments {
		res = append(res, Sub(x, seg)...)
	}

	region.Segments = res
}

func (region *Region) IsEmpty() bool {
	return len(region.Segments) == 0
}

func Sub(seg1 Range, seg2 Range) []Range {
	res := []Range{}

	if seg1.Begin < seg2.Begin && seg1.End > seg2.End {
		res = append(res, Make(seg1.Begin, seg2.Begin))
		res = append(res, Make(seg2.End, seg1.End))
		//res.add(Range.make(seg1.left, seg2.left));
		//res.add(Range.make(seg2.right, seg1.right));
	} else if seg1.End <= seg2.Begin || seg2.End <= seg1.Begin {
		// [ seg1 )
		//          [ seg2 ) -> [   )
		res = append(res, seg1)
	} else if seg2.Begin > seg1.Begin && seg2.Begin < seg1.End {
		// [ seg1 )
		//    [ seg2 ) -> [  )
		res = append(res, Make(seg1.Begin, seg2.Begin))
	} else if seg2.End > seg1.Begin && seg2.End < seg1.End {
		//     [ seg1 )
		// [ seg2 )     -> [  )
		res = append(res, Make(seg2.End, seg1.End))
	}

	return res
}

func (region Region) Begin() uint64 {
	return region.Segments[0].Begin
}

type PieceBlock struct {
	PieceIndex int
	BlockIndex int
}

func FromOffset(offset uint64) PieceBlock {
	return PieceBlock{PieceIndex: int(offset / PIECE_SIZE_UINT64), BlockIndex: int(offset%PIECE_SIZE_UINT64) / BLOCK_SIZE}
}

func (pb PieceBlock) Start() uint64 {
	return uint64(pb.PieceIndex)*PIECE_SIZE_UINT64 + uint64(pb.BlockIndex)*uint64(BLOCK_SIZE)
}
