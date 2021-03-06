package table

import (
	"unsafe"

	"../config"
	"../exe"
	"../index"
	"../memory"
)

var (
	NORMAL_PAGE_SIZE = 1024 * 1024 //Test
)

const (
	PREV_OFFSET   = 8
	NEXT_OFFSET   = 16
	FREE_P_OFFSET = 24
	END_OF_PAGE   = 32
)

func init() {
	m := config.LoadConfig("page")
	if m == nil {
		return
	}
	size, ok := m["page_size"].(float64)
	if !ok {
		return
	}
	NORMAL_PAGE_SIZE = int(size)
}

type Page struct {
	memory.DataBlock
}

// This struct is only to refrence to programer
type page struct {
	table uintptr
	prev  uintptr //DataBlock RawPtr
	next  uintptr //DataBlock RawPtr
	fp    uint
}

func (p *Page) GetTable() *Table {
	ptr, _ := p.Read(0, 8)
	return (*Table)(unsafe.Pointer(uintptr(bytes2uint(ptr))))
}

func (t *Table) NewPage() *Page {
	db, _ := memory.CreateImage(NORMAL_PAGE_SIZE)
	db.Write(0, uint2bytes(uint(uintptr(unsafe.Pointer(t)))))
	db.Write(PREV_OFFSET, uint2bytes(0))
	db.Write(NEXT_OFFSET, uint2bytes(0))
	//	db.Write(END_OF_PAGE, uint2bytes(64)) //BUG
	db.Write(FREE_P_OFFSET, uint2bytes(64))
	return &Page{
		DataBlock: *db,
	}
}

func (p *Page) NextPage() *Page {
	ptr, _ := p.Read(NEXT_OFFSET, 8)
	if bytes2uint(ptr) == 0 {
		return nil
	}
	return &Page{
		DataBlock: *(memory.DataBlockTable[uintptr(bytes2uint(ptr))]),
	}
}

func (p *Page) PrevPage() *Page {
	ptr, _ := p.Read(PREV_OFFSET, 8)
	if bytes2uint(ptr) == 0 {
		return nil
	}
	return &Page{
		DataBlock: *(memory.DataBlockTable[uintptr(bytes2uint(ptr))]),
	}
}

func (p *Page) AppendPage() {
	t := p.GetTable()
	pNew := t.NewPage()
	p.Write(NEXT_OFFSET, uint2bytes(uint(uintptr(pNew.RawPtr))))
	pNew.Write(PREV_OFFSET, uint2bytes(uint(uintptr(p.RawPtr))))
	if t.LastPage == p {
		t.LastPage = pNew
	}
}

func (p *Page) GetFreePos() uint {
	data, _ := p.Read(FREE_P_OFFSET, 8)
	return bytes2uint(data)
}

//func (p *Page) GetEOP() uint {
//	data, _ := p.Read(END_OF_PAGE, 8)
//	return bytes2uint(data)
//}

func (p *Page) ForwardFreePos(i uint) {
	fp := p.GetFreePos()
	fp += i
	p.Write(FREE_P_OFFSET, uint2bytes(fp))
}

func (p *Page) BackFreePos(i uint) {
	fp := p.GetFreePos()
	fp -= i
	p.Write(FREE_P_OFFSET, uint2bytes(fp))
}

func (p *Page) SetFreePos(i uint) {
	p.Write(FREE_P_OFFSET, uint2bytes(i))
}

func (p *Page) AppendField(f *Field, data []byte) {
	if data == nil || len(data) == 0 {
		switch f.Type {
		case exe.INT, exe.FLOAT, exe.OBJECT, exe.ARRAY:
			data = make([]byte, 8)
		case exe.STRING:
			data = make([]byte, 0)
		}
	}
	if !f.FixedSize {
		p.Append(uint322bytes(uint32(len(data))))
	}
	if f.Index != nil {
		ptr := uint(p.RawPtr)
		ptr <<= 24
		fp := p.GetFreePos()
		ptr |= fp
		f.Index.I.Insert(index.BKDRHash(data), uintptr(ptr))
	}
	p.Append(data)
}

func (p *Page) Append(data []byte) {
	fp := p.GetFreePos()
	p.Write(fp, data)
	fp += uint(len(data))
	p.Write(FREE_P_OFFSET, uint2bytes(fp))
	p.Write(END_OF_PAGE, uint2bytes(fp))
}
