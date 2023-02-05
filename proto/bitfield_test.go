package proto

import "testing"

func Test_utils(t *testing.T) {
	if BitsToBytes(28) != 4 {
		t.Errorf("BitsToBytes wrong!")
	}
}

func Test_serialize(t *testing.T) {
	data := []byte{0x0c, 0x00, 0x0f, 0x70}
	sb := StateBuffer{Data: data}
	var bf BitField
	sb.Read(&bf)
	template := []bool{false, false, false, false, true, true, true, true, false, true, true, true}
	if sb.Error() != nil {
		t.Errorf("Unable to read BitField %v", sb.Error())
	} else {
		if bf.Bits() != len(template) {
			t.Errorf("BitField count %d not match %d", bf.Bits(), len(template))
		} else {
			for i := 0; i < len(template); i++ {
				if bf.GetBit(i) != template[i] {
					t.Errorf("BitField not match in %d expected %v", i, template[i])
				}
			}
		}
	}
}

func Test_empty(t *testing.T) {
	data := []byte{0x00, 0x00}
	sb := StateBuffer{Data: data}
	var bf BitField
	sb.Read(&bf)
	if sb.Error() != nil {
		t.Errorf("Unable to read BitField %v", sb.Error())
	} else {
		if bf.Bits() != 0 {
			t.Errorf("BitField size %d is wrong", bf.Bits())
		}

		res := []byte{0x0, 0x0}
		sbRes := StateBuffer{Data: res}
		sbRes.Write(&bf)
		if sbRes.Error() != nil {
			t.Errorf("Unable to write bit field %v", sbRes.Error())
		}
	}

}

func Test_writing(t *testing.T) {
	data := []byte{0x08, 0x07, 0x0a, 0xff}
	var bf BitField
	bf.Assign(data, 28)
	if bf.Size() != 6 {
		t.Errorf("BitField size is not 6: %v", bf.Size())
	}
	res := []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0}
	sb := StateBuffer{Data: res}
	bf.Put(&sb)
	if sb.Error() != nil {
		t.Errorf("Error write BitField %v", sb.Error())
	} else {
		sbR := StateBuffer{Data: res[2:]}
		val := sbR.ReadUint32()
		if sbR.Error() != nil {
			t.Errorf("Unable to read uint32 from res buffer %v", sbR.Error())
		} else {
			if val != uint32(0xf00a0708) {
				t.Errorf("Write incorrect bytes from BitField: %x", val)
			}
		}
	}
}

func Test_tail(t *testing.T) {
	content := []byte{0x0, 0xff}
	bf := BitField{}
	bf.Assign(content, 12)
	if bf.Count() != 4 {
		t.Errorf("Count mot match %d", bf.Count())
	}

	if bf.GetBit(11) != true || bf.GetBit(10) != true || bf.GetBit(9) != true || bf.GetBit(8) != true || bf.GetBit(6) != false {
		t.Errorf("Wrong bits")
	}
}

func Test_adv(t *testing.T) {
	content := []byte{7, 11, 16}
	bf := BitField{}
	bf.Assign(content, 22)
	if bf.Count() != 7 {
		t.Errorf("Count is not correct %d", bf.Count())
	}

	bf.ResizeVal(24, true)

	if bf.Count() != 9 {
		t.Errorf("After resize count is not correct %d", bf.Count())
	}

	if bf.Size() != 5 {
		t.Errorf("Size is not correct %d", bf.Size())
	}

	bf.ResizeVal(20, false)

	if bf.Count() != 7 {
		t.Errorf("Count is not correct %d", bf.Count())
	}

	if bf.Size() != 5 {
		t.Errorf("Size is not correct %d", bf.Size())
	}

	bf.ResizeVal(25, false)

	if bf.Count() != 7 {
		t.Errorf("Count is not correct %d", bf.Count())
	}

	if bf.Size() != 6 {
		t.Errorf("Size is not correct %d", bf.Size())
	}
}
