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
