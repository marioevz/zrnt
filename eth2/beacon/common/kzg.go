package common

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"

	blsu "github.com/protolambda/bls12-381-util"
	"github.com/protolambda/ztyp/codec"
	"github.com/protolambda/ztyp/tree"
	. "github.com/protolambda/ztyp/view"

	"github.com/protolambda/zrnt/eth2/util/hashing"
)

const KZGCommitmentSize = 48
const KZGProofSize = 48

type KZGCommitment [KZGCommitmentSize]byte

var KZGCommitmentType = BasicVectorType(ByteType, KZGCommitmentSize)

func (p *KZGCommitment) Deserialize(dr *codec.DecodingReader) error {
	if p == nil {
		return errors.New("nil kzg commitment")
	}
	_, err := dr.Read(p[:])
	return err
}

func (p *KZGCommitment) Serialize(w *codec.EncodingWriter) error {
	return w.Write(p[:])
}

func (KZGCommitment) ByteLength() uint64 {
	return KZGCommitmentSize
}

func (KZGCommitment) FixedLength() uint64 {
	return KZGCommitmentSize
}

func (p KZGCommitment) HashTreeRoot(hFn tree.HashFn) tree.Root {
	var a, b tree.Root
	copy(a[:], p[0:32])
	copy(b[:], p[32:48])
	return hFn(a, b)
}

func (p KZGCommitment) HashTreeProof(hFn tree.HashFn, index tree.Gindex) []Root {
	var a, b tree.Root
	copy(a[:], p[0:32])
	copy(b[:], p[32:48])

	return hFn.HashTreeProof(index, a, b)
}

func (p KZGCommitment) MarshalText() ([]byte, error) {
	return []byte("0x" + hex.EncodeToString(p[:])), nil
}

func (p KZGCommitment) String() string {
	return "0x" + hex.EncodeToString(p[:])
}

func (p *KZGCommitment) UnmarshalText(text []byte) error {
	if p == nil {
		return errors.New("cannot decode into nil KZGCommitment")
	}
	if len(text) >= 2 && text[0] == '0' && (text[1] == 'x' || text[1] == 'X') {
		text = text[2:]
	}
	if len(text) != 2*KZGCommitmentSize {
		return fmt.Errorf("unexpected length string '%s'", string(text))
	}
	_, err := hex.Decode(p[:], text)
	return err
}

func (p *KZGCommitment) ToPubkey() (*blsu.Pubkey, error) {
	var pub blsu.Pubkey
	if err := pub.Deserialize((*[KZGCommitmentSize]byte)(p)); err != nil {
		return nil, err
	}
	return &pub, nil
}

func (p KZGCommitment) ToVersionedHash() (out Hash32) {
	out = hashing.Hash(p[:])
	out[0] = VERSIONED_HASH_VERSION_KZG
	return out
}

func (p KZGCommitment) View() (*KZGCommitmentView, error) {
	dec := codec.NewDecodingReader(bytes.NewReader(p[:]), KZGCommitmentSize)
	return AsKZGCommitment(KZGCommitmentType.Deserialize(dec))
}

type KZGCommitmentView struct {
	*BasicVectorView
}

func AsKZGCommitment(v View, err error) (*KZGCommitmentView, error) {
	c, err := AsBasicVector(v, err)
	return &KZGCommitmentView{c}, err
}

type KZGCommitments []KZGCommitment

func (li *KZGCommitments) Deserialize(spec *Spec, dr *codec.DecodingReader) error {
	return dr.List(func() codec.Deserializable {
		i := len(*li)
		*li = append(*li, KZGCommitment{})
		return &((*li)[i])
	}, KZGCommitmentSize, uint64(spec.MAX_BLOB_COMMITMENTS_PER_BLOCK))
}

func (li KZGCommitments) Serialize(_ *Spec, w *codec.EncodingWriter) error {
	return w.List(func(i uint64) codec.Serializable {
		return &li[i]
	}, KZGCommitmentSize, uint64(len(li)))
}

func (li KZGCommitments) ByteLength(_ *Spec) (out uint64) {
	return KZGCommitmentSize * uint64(len(li))
}

func (*KZGCommitments) FixedLength(*Spec) uint64 {
	return 0
}

func (li KZGCommitments) HashTreeRoot(spec *Spec, hFn tree.HashFn) Root {
	length := uint64(len(li))
	return hFn.ComplexListHTR(func(i uint64) tree.HTR {
		if i < length {
			return &li[i]
		}
		return nil
	}, length, uint64(spec.MAX_BLOB_COMMITMENTS_PER_BLOCK))
}

func (li KZGCommitments) HashTreeProof(spec *Spec, hFn tree.HashFn, index tree.Gindex) []Root {
	length := uint64(len(li))
	return hFn.ComplexListHTP(func(i uint64) tree.HTP {
		if i < length {
			return &li[i]
		}
		return nil
	}, length, uint64(spec.MAX_BLOB_COMMITMENTS_PER_BLOCK), index)
}

func KZGCommitmentsType(spec *Spec) *ComplexListTypeDef {
	return ComplexListType(KZGCommitmentType, uint64(spec.MAX_BLOB_COMMITMENTS_PER_BLOCK))
}

func (li KZGCommitments) View(spec *Spec) (*KZGCommitmentsView, error) {
	typ := KZGCommitmentsType(spec)
	var buf bytes.Buffer
	if err := li.Serialize(spec, codec.NewEncodingWriter(&buf)); err != nil {
		return nil, err
	}
	data := buf.Bytes()
	dec := codec.NewDecodingReader(bytes.NewReader(data), li.ByteLength(nil))
	return AsKZGCommitments(typ.Deserialize(dec))
}

func AsKZGCommitments(v View, err error) (*KZGCommitmentsView, error) {
	c, err := AsComplexList(v, err)
	return &KZGCommitmentsView{c}, err
}

type KZGCommitmentsView struct {
	*ComplexListView
}

func (v *KZGCommitmentsView) Raw(spec *Spec) (*KZGCommitments, error) {
	var buf bytes.Buffer
	if err := v.Serialize(codec.NewEncodingWriter(&buf)); err != nil {
		return nil, err
	}
	if buf.Len()%KZGCommitmentSize != 0 {
		return nil, fmt.Errorf("invalid length for KZGCommitments: %d", buf.Len())
	}
	commitmentCount := buf.Len() / KZGCommitmentSize
	if commitmentCount > int(spec.MAX_BLOB_COMMITMENTS_PER_BLOCK) {
		return nil, fmt.Errorf("too many KZGCommitments: %d", commitmentCount)
	}
	bufBytes := buf.Bytes()
	out := make(KZGCommitments, len(bufBytes)/KZGCommitmentSize)
	for c := 0; c < len(bufBytes)/KZGCommitmentSize; c++ {
		copy(out[c][:], bufBytes[c*KZGCommitmentSize:(c+1)*KZGCommitmentSize])
	}
	return &out, nil
}

type KZGProof [KZGProofSize]byte

var KZGProofType = BasicVectorType(ByteType, KZGProofSize)

func (p *KZGProof) Deserialize(dr *codec.DecodingReader) error {
	if p == nil {
		return errors.New("nil kzg proof")
	}
	_, err := dr.Read(p[:])
	return err
}

func (p *KZGProof) Serialize(w *codec.EncodingWriter) error {
	return w.Write(p[:])
}

func (KZGProof) ByteLength() uint64 {
	return KZGProofSize
}

func (KZGProof) FixedLength() uint64 {
	return KZGProofSize
}

func (p KZGProof) HashTreeRoot(hFn tree.HashFn) tree.Root {
	var a, b tree.Root
	copy(a[:], p[0:32])
	copy(b[:], p[32:48])
	return hFn(a, b)
}

func (p KZGProof) HashTreeProof(hFn tree.HashFn, index tree.Gindex) []tree.Root {
	var a, b tree.Root
	copy(a[:], p[0:32])
	copy(b[:], p[32:48])
	return hFn.HashTreeProof(index, a, b)
}

func (p KZGProof) MarshalText() ([]byte, error) {
	return []byte("0x" + hex.EncodeToString(p[:])), nil
}

func (p KZGProof) String() string {
	return "0x" + hex.EncodeToString(p[:])
}

func (p *KZGProof) UnmarshalText(text []byte) error {
	if p == nil {
		return errors.New("cannot decode into nil KZGProof")
	}
	if len(text) >= 2 && text[0] == '0' && (text[1] == 'x' || text[1] == 'X') {
		text = text[2:]
	}
	if len(text) != 2*KZGProofSize {
		return fmt.Errorf("unexpected length string '%s'", string(text))
	}
	_, err := hex.Decode(p[:], text)
	return err
}

func (p KZGProof) View() (*KZGProofView, error) {
	dec := codec.NewDecodingReader(bytes.NewReader(p[:]), KZGProofSize)
	return AsKZGProof(KZGProofType.Deserialize(dec))
}

type KZGProofView struct {
	*BasicVectorView
}

func AsKZGProof(v View, err error) (*KZGProofView, error) {
	c, err := AsBasicVector(v, err)
	return &KZGProofView{c}, err
}

type KZGProofs []KZGProof

func (li *KZGProofs) Deserialize(spec *Spec, dr *codec.DecodingReader) error {
	return dr.List(func() codec.Deserializable {
		i := len(*li)
		*li = append(*li, KZGProof{})
		return &((*li)[i])
	}, KZGProofSize, uint64(spec.MAX_BLOB_COMMITMENTS_PER_BLOCK))
}

func (li KZGProofs) Serialize(_ *Spec, w *codec.EncodingWriter) error {
	return w.List(func(i uint64) codec.Serializable {
		return &li[i]
	}, KZGProofSize, uint64(len(li)))
}

func (li KZGProofs) ByteLength(_ *Spec) (out uint64) {
	return KZGProofSize * uint64(len(li))
}

func (*KZGProofs) FixedLength(*Spec) uint64 {
	return 0
}

func (li KZGProofs) HashTreeRoot(spec *Spec, hFn tree.HashFn) Root {
	length := uint64(len(li))
	return hFn.ComplexListHTR(func(i uint64) tree.HTR {
		if i < length {
			return &li[i]
		}
		return nil
	}, length, uint64(spec.MAX_BLOB_COMMITMENTS_PER_BLOCK))
}

func (li KZGProofs) HashTreeProof(spec *Spec, hFn tree.HashFn, index tree.Gindex) []Root {
	length := uint64(len(li))
	return hFn.ComplexListHTP(func(i uint64) tree.HTP {
		if i < length {
			return &li[i]
		}
		return nil
	}, length, uint64(spec.MAX_BLOB_COMMITMENTS_PER_BLOCK), index)
}

func KZGProofsType(spec *Spec) *ComplexListTypeDef {
	return ComplexListType(KZGProofType, uint64(spec.MAX_BLOB_COMMITMENTS_PER_BLOCK))
}

func (li KZGProofs) View(spec *Spec) (*KZGProofsView, error) {
	typ := KZGProofsType(spec)
	var buf bytes.Buffer
	if err := li.Serialize(spec, codec.NewEncodingWriter(&buf)); err != nil {
		return nil, err
	}
	data := buf.Bytes()
	dec := codec.NewDecodingReader(bytes.NewReader(data), li.ByteLength(nil))
	return AsKZGProofs(typ.Deserialize(dec))
}

func AsKZGProofs(v View, err error) (*KZGProofsView, error) {
	c, err := AsComplexList(v, err)
	return &KZGProofsView{c}, err
}

type KZGProofsView struct {
	*ComplexListView
}

func (v *KZGProofsView) Raw(spec *Spec) (*KZGProofs, error) {
	var buf bytes.Buffer
	if err := v.Serialize(codec.NewEncodingWriter(&buf)); err != nil {
		return nil, err
	}
	if buf.Len()%KZGProofSize != 0 {
		return nil, fmt.Errorf("invalid length for KZGProofs: %d", buf.Len())
	}
	commitmentCount := buf.Len() / KZGProofSize
	if commitmentCount > int(spec.MAX_BLOB_COMMITMENTS_PER_BLOCK) {
		return nil, fmt.Errorf("too many KZGProofs: %d", commitmentCount)
	}
	bufBytes := buf.Bytes()
	out := make(KZGProofs, len(bufBytes)/KZGProofSize)
	for c := 0; c < len(bufBytes)/KZGProofSize; c++ {
		copy(out[c][:], bufBytes[c*KZGProofSize:(c+1)*KZGProofSize])
	}
	return &out, nil
}
