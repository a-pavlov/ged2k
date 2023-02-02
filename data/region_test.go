package data

import "testing"

func Test_IntersectSubsSubs(t *testing.T) {
	rg := MakeRegion(Range{Begin: 100, End: 200})
	rg.Sub(Range{150, 180})

	if len(rg.Segments) != 2 {
		t.Errorf("Segments count does not match: %d", len(rg.Segments))
	} else {
		s1 := Range{Begin: 100, End: 150}
		if rg.Segments[0] != s1 {
			t.Error("First region not match")
		}

		s2 := Range{Begin: 180, End: 200}
		if rg.Segments[1] != s2 {
			t.Error("Second segment does not match")
		}

		rg.Sub(Range{Begin: 120, End: 130})
		rg.Sub(Range{Begin: 180, End: 190})
		if len(rg.Segments) != 3 {
			t.Errorf("Segments count does not match: %d round 2", len(rg.Segments))
		}

		s21 := Range{Begin: 100, End: 120}
		s22 := Range{Begin: 130, End: 150}
		s23 := Range{Begin: 190, End: 200}

		if rg.Segments[0] != s21 || rg.Segments[1] != s22 || rg.Segments[2] != s23 {
			t.Error("Wrong segments")
		}
	}

	//assertEquals(rg, new Region(new Range[] {new Range(100L, 150L), new Range(180L, 200L)}));
	//rg.sub(Range.make(120L, 130L)).sub(Range.make(180L, 190L));
	//assertThat(rg, is(new Region(new Range[] {Range.make(100L, 120L), Range.make(130L, 150L), Range.make(190L, 200L)})));
}

func Test_NonIntersectSub(t *testing.T) {
	rg := MakeRegion(Range{Begin: 100, End: 200})
	rg.Sub(Range{Begin: 0, End: 100})
	rg.Sub(Range{Begin: 201, End: 220})
	if len(rg.Segments) != 1 {
		t.Error("Wrong segments count")
	} else {
		s := Range{Begin: 100, End: 200}
		if rg.Segments[0] != s {
			t.Error("Wrong region")
		}
	}
}

func Test_PartialIntersectSub(t *testing.T) {
	rg := MakeRegion(Range{Begin: 100, End: 200})
	rg.Sub(Range{Begin: 0, End: 110})
	rg.Sub(Range{Begin: 190, End: 220})
	if len(rg.Segments) != 1 {
		t.Error("Wrong segments count")
	} else {
		s := Range{Begin: 110, End: 190}
		if rg.Segments[0] != s {
			t.Error("Wrong region result")
		}
	}
}

func Test_FullIntersect(t *testing.T) {
	rg := MakeRegion(Range{Begin: 0, End: 40})
	rg.Sub(Range{Begin: 0, End: 10})
	rg.Sub(Range{Begin: 10, End: 40})
	if len(rg.Segments) != 0 {
		t.Errorf("segments count is not correct %d", len(rg.Segments))
	}
}