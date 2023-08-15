package deneb

import (
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/protolambda/zrnt/eth2/beacon/common"
	"github.com/protolambda/ztyp/codec"
	"github.com/protolambda/ztyp/tree"
	"github.com/protolambda/ztyp/view"
)

const FIELD_ELEMENTS_PER_BLOB = 4096
const BYTES_PER_FIELD_ELEMENT = 32
const BlobSize = FIELD_ELEMENTS_PER_BLOB * BYTES_PER_FIELD_ELEMENT

type Blob [BlobSize]byte

var BlobType = view.BasicVectorType(view.ByteType, BlobSize)

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

func BlobsType(spec *common.Spec) *view.ComplexListTypeDef {
	return view.ComplexListType(BlobType, uint64(spec.MAX_BLOBS_PER_BLOCK))
}

type BlobsBundle struct {
	KZGCommitments KZGCommitments `json:"kzg_commitments" yaml:"kzg_commitments"`
	KZGProofs      KZGProofs      `json:"kzg_proofs" yaml:"kzg_proofs"`
	Blobs          Blobs          `json:"blobs" yaml:"blobs"`
}
