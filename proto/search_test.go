package proto

import (
	"testing"
)

func Test__searchCommon(t *testing.T) {
	bracket_expr := []string{
		"(a b)c d",
		"(a AND b) AND c d",
		"(a b) c AND d",
		"(((a b)))c d",
		"(((a b)))(c)(d)",
		"(((a AND b)))AND((c))AND((d))",
		"(((\"a\" AND \"b\")))AND((c))AND((\"d\"))",
		"   (   (  (  a    AND b   )  )   )  AND  ((c  )  )    AND (  (  d  )   )"}

	for _, s := range bracket_expr {
		parsed, err := BuildEntries(0, 0, 0, 0, "", "", "", 0, 0, s)
		if err != nil {
			t.Errorf("Compile search expression failed with error %v", err)
		} else {
			_, err_p := PackRequest(parsed)
			if err_p != nil {
				t.Errorf("Pack request error %v", err_p)
			}
		}
	}
}

func Test_largeExpr(t *testing.T) {
	parsed, err := BuildEntries(0, 0, 0, 0, "", "", "", 0, 0, "a OR (b OR c AND d OR e) OR j (x OR (y z))")
	if err != nil {
		t.Errorf("Large expression building failed with %v", err)
	} else {
		if len(parsed) != 23 {
			t.Errorf("Parsed elements count incorrect %d expected 23", len(parsed))
		} else {
			entries, err_p := PackRequest(parsed)
			if err_p != nil {
				t.Errorf("Pack request error %v", err_p)
			} else if len(entries) != 17 {
				t.Errorf("Generated entries count %d, expected 17", len(entries))
			} else {
				_, ok_0 := entries[0].(*OperatorEntry)
				_, ok_2 := entries[2].(*OperatorEntry)
				_, ok_3 := entries[3].(*OperatorEntry)
				_, ok_5 := entries[5].(*OperatorEntry)
				_, ok_7 := entries[7].(*OperatorEntry)
				_, ok_10 := entries[10].(*OperatorEntry)
				_, ok_12 := entries[12].(*OperatorEntry)
				_, ok_14 := entries[14].(*OperatorEntry)

				if !ok_0 || !ok_2 || !ok_3 || !ok_5 || !ok_7 || !ok_10 || !ok_12 || !ok_14 {
					t.Error("Operator places error")
				}

				s_1, ok_1 := entries[1].(*StringEntry)
				s_13, ok_13 := entries[13].(*StringEntry)
				s_15, ok_15 := entries[15].(*StringEntry)
				s_16, ok_16 := entries[16].(*StringEntry)

				if !ok_1 || !ok_13 || !ok_15 || !ok_16 {
					t.Error("String places failed")
				} else if string(s_1.value) != "a" || string(s_13.value) != "x" || string(s_15.value) != "y" || string(s_16.value) != "z" {
					t.Error("Wrong string values")
				}

				/*
					assertTrue(sr.entry(0) instanceof BooleanEntry && ((BooleanEntry)sr.entry(0)).operator() == Operator.OPER_OR);
					assertTrue(sr.entry(1) instanceof StringEntry && sr.entry(1).toString().compareTo("a") == 0);
					assertTrue(sr.entry(2) instanceof BooleanEntry && ((BooleanEntry)sr.entry(2)).operator() == Operator.OPER_OR);
					assertTrue(sr.entry(3) instanceof BooleanEntry && ((BooleanEntry)sr.entry(3)).operator() == Operator.OPER_OR);
					assertTrue(sr.entry(4) instanceof StringEntry && sr.entry(4).toString().compareTo("b") == 0);
					assertTrue(sr.entry(5) instanceof BooleanEntry && ((BooleanEntry)sr.entry(5)).operator() == Operator.OPER_AND);
					assertTrue(sr.entry(6) instanceof StringEntry && sr.entry(6).toString().compareTo("c") == 0);
					assertTrue(sr.entry(7) instanceof BooleanEntry && ((BooleanEntry)sr.entry(7)).operator() == Operator.OPER_OR);
					assertTrue(sr.entry(8) instanceof StringEntry && sr.entry(8).toString().compareTo("d") == 0);
					assertTrue(sr.entry(9) instanceof StringEntry && sr.entry(9).toString().compareTo("e") == 0);
					assertTrue(sr.entry(10) instanceof BooleanEntry && ((BooleanEntry)sr.entry(10)).operator() == Operator.OPER_AND);
					assertTrue(sr.entry(11) instanceof StringEntry && sr.entry(11).toString().compareTo("j") == 0);
					assertTrue(sr.entry(12) instanceof BooleanEntry && ((BooleanEntry)sr.entry(12)).operator() == Operator.OPER_OR);
					assertTrue(sr.entry(13) instanceof StringEntry && sr.entry(13).toString().compareTo("x") == 0);
					assertTrue(sr.entry(14) instanceof BooleanEntry && ((BooleanEntry)sr.entry(14)).operator() == Operator.OPER_AND);
					assertTrue(sr.entry(15) instanceof StringEntry && sr.entry(15).toString().compareTo("y") == 0);
					assertTrue(sr.entry(16) instanceof StringEntry && sr.entry(16).toString().compareTo("z") == 0);
				*/
			}
		}
	}
}
