package deneb

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/protolambda/zrnt/eth2/beacon/common"
	"github.com/protolambda/ztyp/codec"
	"github.com/protolambda/ztyp/tree"
	. "github.com/protolambda/ztyp/view"
)

const FIELD_ELEMENTS_PER_BLOB = 4096
const BYTES_PER_FIELD_ELEMENT = 32
const BlobSize = FIELD_ELEMENTS_PER_BLOB * BYTES_PER_FIELD_ELEMENT

type Blob [BlobSize]byte

var BlobType = BasicVectorType(ByteType, BlobSize)

func (p *Blob) Deserialize(dr *codec.DecodingReader) error {
	if p == nil {
		return errors.New("nil blob")
	}
	_, err := dr.Read(p[:])
	return err
}

func (p *Blob) Serialize(w *codec.EncodingWriter) error {
	return w.Write(p[:])
}

func (Blob) ByteLength() uint64 {
	return BlobSize
}

func (Blob) FixedLength() uint64 {
	return BlobSize
}

func (p Blob) HashTreeRoot(hFn tree.HashFn) tree.Root {
	return hFn.ByteVectorHTR(p[:])
}

func (p Blob) MarshalText() ([]byte, error) {
	return []byte("0x" + hex.EncodeToString(p[:])), nil
}

func (p Blob) String() string {
	return "0x" + hex.EncodeToString(p[:])
}

func (p *Blob) UnmarshalText(text []byte) error {
	if p == nil {
		return errors.New("cannot decode into nil Blob")
	}
	if len(text) >= 2 && text[0] == '0' && (text[1] == 'x' || text[1] == 'X') {
		text = text[2:]
	}
	if len(text) != 2*BlobSize {
		return fmt.Errorf("unexpected length string '%s'", string(text))
	}
	_, err := hex.Decode(p[:], text)
	return err
}

type Blobs []Blob

func (li *Blobs) Deserialize(spec *common.Spec, dr *codec.DecodingReader) error {
	return dr.List(func() codec.Deserializable {
		i := len(*li)
		*li = append(*li, Blob{})
		return &((*li)[i])
	}, BlobSize, uint64(spec.MAX_BLOBS_PER_BLOCK))
}

func (li Blobs) Serialize(_ *common.Spec, w *codec.EncodingWriter) error {
	return w.List(func(i uint64) codec.Serializable {
		return &li[i]
	}, BlobSize, uint64(len(li)))
}

func (li Blobs) ByteLength(_ *common.Spec) (out uint64) {
	return BlobSize * uint64(len(li))
}

func (*Blobs) FixedLength(*common.Spec) uint64 {
	return 0
}

func (li Blobs) HashTreeRoot(spec *common.Spec, hFn tree.HashFn) common.Root {
	length := uint64(len(li))
	return hFn.ComplexListHTR(func(i uint64) tree.HTR {
		if i < length {
			return &li[i]
		}
		return nil
	}, length, uint64(spec.MAX_BLOBS_PER_BLOCK))
}

func BlobsType(spec *common.Spec) *ComplexListTypeDef {
	return ComplexListType(BlobType, uint64(spec.MAX_BLOBS_PER_BLOCK))
}

func (li Blobs) View(spec *common.Spec) (*BlobsView, error) {
	typ := BlobsType(spec)
	var buf bytes.Buffer
	if err := li.Serialize(spec, codec.NewEncodingWriter(&buf)); err != nil {
		return nil, err
	}
	data := buf.Bytes()
	dec := codec.NewDecodingReader(bytes.NewReader(data), li.ByteLength(nil))
	return AsBlobs(typ.Deserialize(dec))
}

func AsBlobs(v View, err error) (*BlobsView, error) {
	c, err := AsComplexList(v, err)
	return &BlobsView{c}, err
}

type BlobsView struct {
	*ComplexListView
}

func (v *BlobsView) Raw(spec *common.Spec) (*Blobs, error) {
	var buf bytes.Buffer
	if err := v.Serialize(codec.NewEncodingWriter(&buf)); err != nil {
		return nil, err
	}
	if buf.Len()%BlobSize != 0 {
		return nil, fmt.Errorf("invalid length for blobs: %d", buf.Len())
	}
	commitmentCount := buf.Len() / BlobSize
	if commitmentCount > int(spec.MAX_BLOBS_PER_BLOCK) {
		return nil, fmt.Errorf("too many Blobs: %d", commitmentCount)
	}
	bufBytes := buf.Bytes()
	out := make(Blobs, len(bufBytes)/BlobSize)
	for c := 0; c < len(bufBytes)/BlobSize; c++ {
		copy(out[c][:], bufBytes[c*BlobSize:(c+1)*BlobSize])
	}
	return &out, nil
}
