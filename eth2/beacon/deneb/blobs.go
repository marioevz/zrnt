package deneb

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/protolambda/zrnt/eth2/beacon/common"
	"github.com/protolambda/ztyp/codec"
	"github.com/protolambda/ztyp/tree"
	. "github.com/protolambda/ztyp/view"
)

const BYTES_PER_FIELD_ELEMENT = 32

// Eth1 deposit ordering
type BlobIndex Uint64View

func AsBlobIndex(v View, err error) (BlobIndex, error) {
	i, err := AsUint64(v, err)
	return BlobIndex(i), err
}

func (i *BlobIndex) Deserialize(dr *codec.DecodingReader) error {
	return (*Uint64View)(i).Deserialize(dr)
}

func (i BlobIndex) Serialize(w *codec.EncodingWriter) error {
	return w.WriteUint64(uint64(i))
}

func (BlobIndex) ByteLength() uint64 {
	return 8
}

func (BlobIndex) FixedLength() uint64 {
	return 8
}

func (i BlobIndex) HashTreeRoot(hFn tree.HashFn) tree.Root {
	return Uint64View(i).HashTreeRoot(hFn)
}

func (e BlobIndex) MarshalJSON() ([]byte, error) {
	return Uint64View(e).MarshalJSON()
}

func (e *BlobIndex) UnmarshalJSON(b []byte) error {
	return ((*Uint64View)(e)).UnmarshalJSON(b)
}

func (e BlobIndex) String() string {
	return Uint64View(e).String()
}

type Blob []byte

func BlobSize(spec *common.Spec) uint64 {
	return uint64(spec.FIELD_ELEMENTS_PER_BLOB) * BYTES_PER_FIELD_ELEMENT
}

func BlobType(spec *common.Spec) *BasicVectorTypeDef {
	return BasicVectorType(ByteType, BlobSize(spec))
}

func (p *Blob) Deserialize(spec *common.Spec, dr *codec.DecodingReader) error {
	if p == nil {
		return errors.New("nil blob")
	}
	*p = make(Blob, BlobSize(spec))
	n, err := dr.Read((*p)[:])
	if err != nil {
		return err
	}
	if uint64(n) != BlobSize(spec) {
		return errors.New("incorrect number of bytes read")
	}
	return nil
}

func (p *Blob) Serialize(spec *common.Spec, w *codec.EncodingWriter) error {
	return w.Write((*p)[:])
}

func (Blob) ByteLength(spec *common.Spec) uint64 {
	return spec.FIELD_ELEMENTS_PER_BLOB.ByteLength()
}

func (Blob) FixedLength(spec *common.Spec) uint64 {
	return spec.FIELD_ELEMENTS_PER_BLOB.ByteLength()
}

func (p Blob) HashTreeRoot(spec *common.Spec, hFn tree.HashFn) tree.Root {
	return hFn.ByteVectorHTR(p[:])
}

/*
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
	if len(text) != int(2*BlobSize(spec)) {
		return fmt.Errorf("unexpected length string '%s'", string(text))
	}
	_, err := hex.Decode((*p)[:], text)
	return err
}
*/

type Blobs []Blob

func (li *Blobs) Deserialize(spec *common.Spec, dr *codec.DecodingReader) error {
	return dr.List(func() codec.Deserializable {
		i := len(*li)
		*li = append(*li, Blob{})
		return spec.Wrap(&((*li)[i]))
	}, BlobSize(spec), uint64(spec.MAX_BLOBS_PER_BLOCK))
}

func (li Blobs) Serialize(spec *common.Spec, w *codec.EncodingWriter) error {
	return w.List(func(i uint64) codec.Serializable {
		return spec.Wrap(&li[i])
	}, BlobSize(spec), uint64(len(li)))
}

func (li Blobs) ByteLength(spec *common.Spec) (out uint64) {
	return BlobSize(spec) * uint64(len(li))
}

func (*Blobs) FixedLength(*common.Spec) uint64 {
	return 0
}

func (li Blobs) HashTreeRoot(spec *common.Spec, hFn tree.HashFn) common.Root {
	length := uint64(len(li))
	return hFn.ComplexListHTR(func(i uint64) tree.HTR {
		if i < length {
			return spec.Wrap(&li[i])
		}
		return nil
	}, length, uint64(spec.MAX_BLOBS_PER_BLOCK))
}

func (li Blobs) Roots(spec *common.Spec, hFn tree.HashFn) BlobRoots {
	length := uint64(len(li))
	roots := make(BlobRoots, length)
	for i := uint64(0); i < length; i++ {
		roots[i] = li[i].HashTreeRoot(spec, hFn)
	}
	return roots
}

func BlobsType(spec *common.Spec) *ComplexListTypeDef {
	return ComplexListType(BlobType(spec), uint64(spec.MAX_BLOBS_PER_BLOCK))
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
	blobSize := int(BlobSize(spec))
	if buf.Len()%blobSize != 0 {
		return nil, fmt.Errorf("invalid length for blobs: %d", buf.Len())
	}
	commitmentCount := buf.Len() / blobSize
	if commitmentCount > int(spec.MAX_BLOBS_PER_BLOCK) {
		return nil, fmt.Errorf("too many Blobs: %d", commitmentCount)
	}
	bufBytes := buf.Bytes()
	out := make(Blobs, len(bufBytes)/blobSize)
	for c := 0; c < len(bufBytes)/blobSize; c++ {
		copy(out[c][:], bufBytes[c*blobSize:(c+1)*blobSize])
	}
	return &out, nil
}

func BlobRootsType(spec *common.Spec) ListTypeDef {
	return ListType(RootType, uint64(spec.MAX_BLOBS_PER_BLOCK))
}

type BlobRoots []tree.Root

func (br *BlobRoots) Deserialize(spec *common.Spec, dr *codec.DecodingReader) error {
	return dr.List(func() codec.Deserializable {
		i := len(*br)
		*br = append(*br, tree.Root{})
		return &(*br)[i]
	}, RootType.TypeByteLength(), uint64(spec.MAX_BLOBS_PER_BLOCK))
}

func (br BlobRoots) Serialize(spec *common.Spec, w *codec.EncodingWriter) error {
	return w.List(func(i uint64) codec.Serializable {
		return &br[i]
	}, RootType.TypeByteLength(), uint64(len(br)))
}

func (br BlobRoots) ByteLength(spec *common.Spec) uint64 {
	return uint64(len(br)) * RootType.TypeByteLength()
}

func (br BlobRoots) FixedLength(spec *common.Spec) uint64 {
	return 0 // it's a list, no fixed length
}

func (br BlobRoots) HashTreeRoot(spec *common.Spec, hFn tree.HashFn) tree.Root {
	return hFn.ChunksHTR(func(i uint64) tree.Root {
		return br[i]
	}, uint64(len(br)), uint64(spec.MAX_BLOBS_PER_BLOCK))
}

func BlobSidecarType(spec *common.Spec) *ContainerTypeDef {
	return ContainerType("BlobSidecar", []FieldDef{
		{"block_root", RootType},
		{"index", Uint64Type},
		{"slot", common.SlotType},
		{"block_parent_root", RootType},
		{"proposer_index", common.ValidatorIndexType},
		{"blob", BlobType(spec)},
		{"kzg_commitment", common.KZGCommitmentType},
		{"kzg_proof", common.KZGProofType},
	})
}

type BlobSidecar struct {
	BlockRoot       common.Root           `json:"block_root" yaml:"block_root"`
	Index           BlobIndex             `json:"index" yaml:"index"`
	Slot            common.Slot           `json:"slot" yaml:"slot"`
	BlockParentRoot common.Root           `json:"block_parent_root" yaml:"block_parent_root"`
	ProposerIndex   common.ValidatorIndex `json:"proposer_index" yaml:"proposer_index"`
	Blob            Blob                  `json:"blob" yaml:"blob"`
	KZGCommitment   common.KZGCommitment  `json:"kzg_commitment" yaml:"kzg_commitment"`
	KZGProof        common.KZGProof       `json:"kzg_proof" yaml:"kzg_proof"`
}

func (b *BlobSidecar) Deserialize(spec *common.Spec, dr *codec.DecodingReader) error {
	return dr.Container(&b.BlockRoot, &b.Index, &b.Slot, &b.BlockParentRoot, &b.ProposerIndex, spec.Wrap(&b.Blob), &b.KZGCommitment, &b.KZGProof)
}

func (b *BlobSidecar) Serialize(spec *common.Spec, w *codec.EncodingWriter) error {
	return w.Container(&b.BlockRoot, &b.Index, &b.Slot, &b.BlockParentRoot, &b.ProposerIndex, spec.Wrap(&b.Blob), &b.KZGCommitment, &b.KZGProof)
}

func (b *BlobSidecar) ByteLength(spec *common.Spec) uint64 {
	return codec.ContainerLength(&b.BlockRoot, &b.Index, &b.Slot, &b.BlockParentRoot, &b.ProposerIndex, spec.Wrap(&b.Blob), &b.KZGCommitment, &b.KZGProof)
}

func (b *BlobSidecar) FixedLength(spec *common.Spec) uint64 {
	return codec.ContainerLength(&b.BlockRoot, &b.Index, &b.Slot, &b.BlockParentRoot, &b.ProposerIndex, spec.Wrap(&b.Blob), &b.KZGCommitment, &b.KZGProof)
}

func (b *BlobSidecar) HashTreeRoot(spec *common.Spec, hFn tree.HashFn) common.Root {
	return hFn.HashTreeRoot(&b.BlockRoot, &b.Index, &b.Slot, &b.BlockParentRoot, &b.ProposerIndex, spec.Wrap(&b.Blob), &b.KZGCommitment, &b.KZGProof)
}

type SignedBlobSidecar struct {
	Message   BlobSidecar         `json:"message" yaml:"message"`
	Signature common.BLSSignature `json:"signature" yaml:"signature"`
}

func (b *SignedBlobSidecar) Deserialize(spec *common.Spec, dr *codec.DecodingReader) error {
	return dr.Container(spec.Wrap(&b.Message), &b.Signature)
}

func (b *SignedBlobSidecar) Serialize(spec *common.Spec, w *codec.EncodingWriter) error {
	return w.Container(spec.Wrap(&b.Message), &b.Signature)
}

func (b *SignedBlobSidecar) ByteLength(spec *common.Spec) uint64 {
	return codec.ContainerLength(spec.Wrap(&b.Message), &b.Signature)
}

func (b *SignedBlobSidecar) FixedLength(spec *common.Spec) uint64 {
	return codec.ContainerLength(spec.Wrap(&b.Message), &b.Signature)
}

func (b *SignedBlobSidecar) HashTreeRoot(spec *common.Spec, hFn tree.HashFn) common.Root {
	return hFn.HashTreeRoot(spec.Wrap(&b.Message), &b.Signature)
}

var BlindedBlobSidecarType = ContainerType("BlindedBlobSidecar", []FieldDef{
	{"block_root", RootType},
	{"index", Uint64Type},
	{"slot", common.SlotType},
	{"block_parent_root", RootType},
	{"proposer_index", common.ValidatorIndexType},
	{"blob_root", RootType},
	{"kzg_commitment", common.KZGCommitmentType},
	{"kzg_proof", common.KZGProofType},
})

type BlindedBlobSidecar struct {
	BlockRoot       common.Root           `json:"block_root" yaml:"block_root"`
	Index           BlobIndex             `json:"index" yaml:"index"`
	Slot            common.Slot           `json:"slot" yaml:"slot"`
	BlockParentRoot common.Root           `json:"block_parent_root" yaml:"block_parent_root"`
	ProposerIndex   common.ValidatorIndex `json:"proposer_index" yaml:"proposer_index"`
	BlobRoot        common.Root           `json:"blob_root" yaml:"blob_root"`
	KZGCommitment   common.KZGCommitment  `json:"kzg_commitment" yaml:"kzg_commitment"`
	KZGProof        common.KZGProof       `json:"kzg_proof" yaml:"kzg_proof"`
}

func (b *BlindedBlobSidecar) Deserialize(dr *codec.DecodingReader) error {
	return dr.Container(&b.BlockRoot, &b.Index, &b.Slot, &b.BlockParentRoot, &b.ProposerIndex, &b.BlobRoot, &b.KZGCommitment, &b.KZGProof)
}

func (b *BlindedBlobSidecar) Serialize(w *codec.EncodingWriter) error {
	return w.Container(&b.BlockRoot, &b.Index, &b.Slot, &b.BlockParentRoot, &b.ProposerIndex, &b.BlobRoot, &b.KZGCommitment, &b.KZGProof)
}

func (b *BlindedBlobSidecar) ByteLength() uint64 {
	return codec.ContainerLength(&b.BlockRoot, &b.Index, &b.Slot, &b.BlockParentRoot, &b.ProposerIndex, &b.BlobRoot, &b.KZGCommitment, &b.KZGProof)
}

func (b *BlindedBlobSidecar) FixedLength() uint64 {
	return codec.ContainerLength(&b.BlockRoot, &b.Index, &b.Slot, &b.BlockParentRoot, &b.ProposerIndex, &b.BlobRoot, &b.KZGCommitment, &b.KZGProof)
}

func (b *BlindedBlobSidecar) HashTreeRoot(hFn tree.HashFn) common.Root {
	return hFn.HashTreeRoot(&b.BlockRoot, &b.Index, &b.Slot, &b.BlockParentRoot, &b.ProposerIndex, &b.BlobRoot, &b.KZGCommitment, &b.KZGProof)
}

func (b *BlindedBlobSidecar) Unblind(blob *Blob) *BlobSidecar {
	return &BlobSidecar{
		BlockRoot:       b.BlockRoot,
		Index:           b.Index,
		Slot:            b.Slot,
		BlockParentRoot: b.BlockParentRoot,
		ProposerIndex:   b.ProposerIndex,
		Blob:            *blob,
		KZGCommitment:   b.KZGCommitment,
		KZGProof:        b.KZGProof,
	}
}

var SignedBlindedBlobSidecarType = ContainerType("SignedBlindedBlobSidecar", []FieldDef{
	{"message", BlindedBlobSidecarType},
	{"signature", common.BLSSignatureType},
})

type SignedBlindedBlobSidecar struct {
	Message   BlindedBlobSidecar  `json:"message" yaml:"message"`
	Signature common.BLSSignature `json:"signature" yaml:"signature"`
}

func (b *SignedBlindedBlobSidecar) Deserialize(dr *codec.DecodingReader) error {
	return dr.Container(&b.Message, &b.Signature)
}

func (b *SignedBlindedBlobSidecar) Serialize(w *codec.EncodingWriter) error {
	return w.Container(&b.Message, &b.Signature)
}

func (b *SignedBlindedBlobSidecar) ByteLength() uint64 {
	return codec.ContainerLength(&b.Message, &b.Signature)
}

func (b *SignedBlindedBlobSidecar) FixedLength() uint64 {
	return codec.ContainerLength(&b.Message, &b.Signature)
}

func (b *SignedBlindedBlobSidecar) HashTreeRoot(hFn tree.HashFn) common.Root {
	return hFn.HashTreeRoot(&b.Message, &b.Signature)
}

func (b *SignedBlindedBlobSidecar) Unblind(blob *Blob) *SignedBlobSidecar {
	return &SignedBlobSidecar{
		Message:   *b.Message.Unblind(blob),
		Signature: b.Signature,
	}
}
