package bencode

import (
	"bufio"
	"errors"
	"io"
	"strconv"
)

var (
	ErrNum = errors.New("expect num")
	ErrCol = errors.New("expect colon")
	ErrEpI = errors.New("expect char i")
	ErrEpE = errors.New("expect char e")
	ErrTyp = errors.New("wrong type")
	ErrIvd = errors.New("invalid bencode")
)

type BType byte

const (
	BSTR  BType = 1
	BINT  BType = 2
	BLIST BType = 3
	BDICT BType = 4
)

type BValue interface{}

type BObject struct {
	type_ BType
	val_  BValue
}

func (o *BObject) Str() (string, error) {
	if o.type_ != BSTR {
		return "", ErrTyp
	}
	return o.val_.(string), nil
}

func (o *BObject) Int() (int, error) {
	if o.type_ != BINT {
		return 0, ErrTyp
	}
	return o.val_.(int), nil
}

func (o *BObject) List() ([]*BObject, error) {
	if o.type_ != BLIST {
		return nil, ErrTyp
	}
	return o.val_.([]*BObject), nil
}

func (o *BObject) Dict() (map[string]*BObject, error) {
	if o.type_ != BDICT {
		return nil, ErrTyp
	}
	return o.val_.(map[string]*BObject), nil
}

func EncodeString(w io.Writer, val string) int {
	strLen := len(val)
	bw, ok := w.(*bufio.Writer)
	if !ok {
		bw = bufio.NewWriter(w)
	}

	lenStr := strconv.Itoa(strLen)
	wLen := len(lenStr)
	bw.WriteString(lenStr)

	bw.WriteByte(':')
	wLen++

	bw.WriteString(val)
	wLen += strLen

	if err := bw.Flush(); err != nil {
		return 0
	}

	return wLen
}

func DecodeString(r io.Reader) (val string, err error) {
	br, ok := r.(*bufio.Reader)
	if !ok {
		br = bufio.NewReader(r)
	}

	var numBytes []byte
	var b byte
	var num int

	for {
		if b, err = br.ReadByte(); err != nil {
			return
		}
		// 前缀是否为数字
		if b >= '0' && b <= '9' {
			numBytes = append(numBytes, b)
			continue
		}
		// 中间是否有冒号
		if b == ':' {
			break
		} else {
			err = ErrCol
			return
		}
	}

	numString := string(numBytes)
	if num, err = strconv.Atoi(numString); err != nil {
		return
	}

	buf := make([]byte, num)
	_, err = io.ReadAtLeast(br, buf, num)
	val = string(buf)
	return
}

func EncodeInt(w io.Writer, val int) (wLen int) {
	bw, ok := w.(*bufio.Writer)
	if !ok {
		bw = bufio.NewWriter(w)
	}
	bw.WriteByte('i')
	wLen++
	valStr := strconv.Itoa(val)
	bw.WriteString(valStr)
	wLen += len(valStr)
	bw.WriteByte('e')
	wLen++
	return
}

func DecodeInt(r io.Reader) (val int, err error) {
	br, ok := r.(*bufio.Reader)
	if !ok {
		br = bufio.NewReader(r)
	}

	var b byte
	//左侧是否为i
	if b, _ = br.ReadByte(); b != 'i' {
		err = ErrEpI
		return
	}
	//是否有负号
	if b, _ = br.ReadByte(); b < '0' || b > '9' || b != '-' {
		err = ErrCol
		return
	}

	var numBytes []byte
	numBytes = append(numBytes, b)

	for {
		if b, err = br.ReadByte(); err != nil {
			return
		}
		// 是否为数字
		if b >= '0' && b <= '9' {
			numBytes = append(numBytes, b)
			continue
		}
		// 右侧是否为e
		if b == 'e' {
			break
		} else {
			err = ErrEpE
			return
		}
	}

	numStr := string(numBytes)
	val, err = strconv.Atoi(numStr)
	return
}

func (o *BObject) Bencode(w io.Writer) int {
	bw, ok := w.(*bufio.Writer)
	if !ok {
		bw = bufio.NewWriter(w)
	}

	wLen := 0
	switch o.type_ {
	case BSTR:
		str, _ := o.Str()
		wLen += EncodeString(bw, str)
	case BINT:
		val, _ := o.Int()
		wLen += EncodeInt(bw, val)
	case BLIST:
		bw.WriteByte('l')
		list, _ := o.List()
		for _, elem := range list {
			wLen += elem.Bencode(bw)
		}
		bw.WriteByte('e')
		wLen += 2
	case BDICT:
		bw.WriteByte('d')
		dict, _ := o.Dict()
		for k, v := range dict {
			wLen += EncodeString(bw, k)
			wLen += v.Bencode(bw)
		}
		bw.WriteByte('e')
		wLen += 2
	}

	bw.Flush()
	return wLen
}
