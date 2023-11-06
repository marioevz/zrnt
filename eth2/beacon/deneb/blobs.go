package deneb

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/protolambda/zrnt/eth2/beacon/common"
	"github.com/protolambda/ztyp/codec"
	"github.com/protolambda/ztyp/conv"
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

func (i BlobIndex) HashTreeProof(hFn tree.HashFn, index tree.Gindex) []tree.Root {
	return Uint64View(i).HashTreeProof(hFn, index)
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
	return BlobSize(spec)
}

func (Blob) FixedLength(spec *common.Spec) uint64 {
	return BlobSize(spec)
}

func (p Blob) HashTreeRoot(spec *common.Spec, hFn tree.HashFn) tree.Root {
	return hFn.ByteVectorHTR(p[:])
}

func (p Blob) HashTreeProof(spec *common.Spec, hFn tree.HashFn, index tree.Gindex) []tree.Root {
	return hFn.ByteVectorHTP(p[:], index)
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
	// TODO: This might yield a blob with an out-of-spec length, but there's no way to recover
	// the spec from the context of the unmarshal.
	return conv.DynamicBytesUnmarshalText((*[]byte)(p), text[:])
}

type Blobs []Blob

func (li *Blobs) Deserialize(spec *common.Spec, dr *codec.DecodingReader) error {
	return dr.List(func() codec.Deserializable {
		i := len(*li)
		*li = append(*li, Blob{})
		return spec.Wrap(&((*li)[i]))
	}, BlobSize(spec), uint64(spec.MAX_BLOB_COMMITMENTS_PER_BLOCK))
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
	}, length, uint64(spec.MAX_BLOB_COMMITMENTS_PER_BLOCK))
}

func (li Blobs) HashTreeProof(spec *common.Spec, hFn tree.HashFn, index tree.Gindex) []common.Root {
	length := uint64(len(li))
	return hFn.ComplexListHTP(func(i uint64) tree.HTP {
		if i < length {
			return spec.Wrap(&li[i])
		}
		return nil
	}, length, uint64(spec.MAX_BLOB_COMMITMENTS_PER_BLOCK), index)
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
	return ComplexListType(BlobType(spec), uint64(spec.MAX_BLOB_COMMITMENTS_PER_BLOCK))
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
	if commitmentCount > int(spec.MAX_BLOB_COMMITMENTS_PER_BLOCK) {
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
	return ListType(RootType, uint64(spec.MAX_BLOB_COMMITMENTS_PER_BLOCK))
}

type BlobRoots []tree.Root

func (br *BlobRoots) Deserialize(spec *common.Spec, dr *codec.DecodingReader) error {
	return dr.List(func() codec.Deserializable {
		i := len(*br)
		*br = append(*br, tree.Root{})
		return &(*br)[i]
	}, RootType.TypeByteLength(), uint64(spec.MAX_BLOB_COMMITMENTS_PER_BLOCK))
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
	length := uint64(len(br))
	return hFn.ComplexListHTR(func(i uint64) tree.HTR {
		if i < length {
			return &br[i]
		}
		return nil
	}, length, uint64(spec.MAX_BLOB_COMMITMENTS_PER_BLOCK))
}

func (br BlobRoots) HashTreeProof(spec *common.Spec, hFn tree.HashFn, index tree.Gindex) []tree.Root {
	length := uint64(len(br))
	return hFn.ComplexListHTP(func(i uint64) tree.HTP {
		if i < length {
			return br[i]
		}
		return nil
	}, length, uint64(spec.MAX_BLOB_COMMITMENTS_PER_BLOCK), index)
}

// KZGCommitmentInclusionProof is a Root vector of KZG_COMMITMENT_INCLUSION_PROOF_DEPTH length
type KZGCommitmentInclusionProof []common.Root

func (a *KZGCommitmentInclusionProof) Deserialize(spec *common.Spec, dr *codec.DecodingReader) error {
	return tree.ReadRoots(dr, (*[]common.Root)(a), uint64(spec.KZG_COMMITMENT_INCLUSION_PROOF_DEPTH))
}

func (a KZGCommitmentInclusionProof) Serialize(spec *common.Spec, w *codec.EncodingWriter) error {
	return tree.WriteRoots(w, a)
}

func (a KZGCommitmentInclusionProof) ByteLength(spec *common.Spec) (out uint64) {
	return uint64(spec.KZG_COMMITMENT_INCLUSION_PROOF_DEPTH) * 32
}

func (a *KZGCommitmentInclusionProof) FixedLength(spec *common.Spec) uint64 {
	return uint64(spec.KZG_COMMITMENT_INCLUSION_PROOF_DEPTH) * 32
}

func (li KZGCommitmentInclusionProof) HashTreeRoot(spec *common.Spec, hFn tree.HashFn) common.Root {
	length := uint64(len(li))
	return hFn.ComplexVectorHTR(func(i uint64) tree.HTR {
		if i < length {
			return &li[i]
		}
		return nil
	}, length)
}

func (li KZGCommitmentInclusionProof) HashTreeProof(spec *common.Spec, hFn tree.HashFn, index tree.Gindex) []common.Root {
	length := uint64(len(li))
	return hFn.ComplexVectorHTP(func(i uint64) tree.HTP {
		if i < length {
			return li[i]
		}
		return nil
	}, length, index)
}

// KZGCommitmentInclusionProofs is a KZGCommitmentInclusionProof list of MAX_BLOBS_PER_BLOCK limit length
type KZGCommitmentInclusionProofs []KZGCommitmentInclusionProof

func (li *KZGCommitmentInclusionProofs) Deserialize(spec *common.Spec, dr *codec.DecodingReader) error {
	return dr.List(func() codec.Deserializable {
		i := len(*li)
		*li = append(*li, KZGCommitmentInclusionProof{})
		return spec.Wrap(&((*li)[i]))
	}, uint64(spec.KZG_COMMITMENT_INCLUSION_PROOF_DEPTH)*32, uint64(spec.MAX_BLOBS_PER_BLOCK))
}

func (li KZGCommitmentInclusionProofs) Serialize(spec *common.Spec, w *codec.EncodingWriter) error {
	return w.List(func(i uint64) codec.Serializable {
		return spec.Wrap(&li[i])
	}, uint64(spec.KZG_COMMITMENT_INCLUSION_PROOF_DEPTH)*32, uint64(len(li)))
}

func (li KZGCommitmentInclusionProofs) ByteLength(spec *common.Spec) (out uint64) {
	return uint64(spec.KZG_COMMITMENT_INCLUSION_PROOF_DEPTH) * 32 * uint64(len(li))
}

func (li *KZGCommitmentInclusionProofs) FixedLength(spec *common.Spec) uint64 {
	return 0
}

func (li KZGCommitmentInclusionProofs) HashTreeRoot(spec *common.Spec, hFn tree.HashFn) common.Root {
	length := uint64(len(li))
	return hFn.ComplexVectorHTR(func(i uint64) tree.HTR {
		if i < length {
			return spec.Wrap(&li[i])
		}
		return nil
	}, length)
}

func (li KZGCommitmentInclusionProofs) HashTreeProof(spec *common.Spec, hFn tree.HashFn, index tree.Gindex) []common.Root {
	length := uint64(len(li))
	return hFn.ComplexVectorHTP(func(i uint64) tree.HTP {
		if i < length {
			return spec.Wrap(&li[i])
		}
		return nil
	}, length, index)
}

func KZGCommitmentInclusionProofType(spec *common.Spec) VectorTypeDef {
	return VectorType(common.Bytes32Type, uint64(spec.KZG_COMMITMENT_INCLUSION_PROOF_DEPTH))
}

func BlobSidecarType(spec *common.Spec) *ContainerTypeDef {
	return ContainerType("BlobSidecar", []FieldDef{
		{"index", Uint64Type},
		{"blob", BlobType(spec)},
		{"kzg_commitment", common.KZGCommitmentType},
		{"kzg_proof", common.KZGProofType},
		{"signed_block_header", common.SignedBeaconBlockHeaderType},
		{"kzg_commitment_inclusion_proof", KZGCommitmentInclusionProofType(spec)},
	})
}

type BlobSidecar struct {
	Index                       BlobIndex                      `json:"index" yaml:"index"`
	Blob                        Blob                           `json:"blob" yaml:"blob"`
	KZGCommitment               common.KZGCommitment           `json:"kzg_commitment" yaml:"kzg_commitment"`
	KZGProof                    common.KZGProof                `json:"kzg_proof" yaml:"kzg_proof"`
	SignedBlockHeader           common.SignedBeaconBlockHeader `json:"signed_block_header" yaml:"signed_block_header"`
	KZGCommitmentInclusionProof KZGCommitmentInclusionProof    `json:"kzg_commitment_inclusion_proof" yaml:"kzg_commitment_inclusion_proof"`
}

func (b *BlobSidecar) Deserialize(spec *common.Spec, dr *codec.DecodingReader) error {
	return dr.Container(&b.Index, spec.Wrap(&b.Blob), &b.KZGCommitment, &b.KZGProof, &b.SignedBlockHeader, spec.Wrap(&b.KZGCommitmentInclusionProof))
}

func (b *BlobSidecar) Serialize(spec *common.Spec, w *codec.EncodingWriter) error {
	return w.Container(&b.Index, spec.Wrap(&b.Blob), &b.KZGCommitment, &b.KZGProof, &b.SignedBlockHeader, spec.Wrap(&b.KZGCommitmentInclusionProof))
}

func (b *BlobSidecar) ByteLength(spec *common.Spec) uint64 {
	return codec.ContainerLength(&b.Index, spec.Wrap(&b.Blob), &b.KZGCommitment, &b.KZGProof, &b.SignedBlockHeader, spec.Wrap(&b.KZGCommitmentInclusionProof))
}

func (b *BlobSidecar) FixedLength(spec *common.Spec) uint64 {
	return codec.ContainerLength(&b.Index, spec.Wrap(&b.Blob), &b.KZGCommitment, &b.KZGProof, &b.SignedBlockHeader, spec.Wrap(&b.KZGCommitmentInclusionProof))
}

func (b *BlobSidecar) HashTreeRoot(spec *common.Spec, hFn tree.HashFn) common.Root {
	return hFn.HashTreeRoot(
		&b.Index,
		spec.Wrap(&b.Blob),
		&b.KZGCommitment,
		&b.KZGProof,
		&b.SignedBlockHeader,
		spec.Wrap(&b.KZGCommitmentInclusionProof),
	)
}

func (b *BlobSidecar) HashTreeProof(spec *common.Spec, hFn tree.HashFn, index tree.Gindex) []common.Root {
	return hFn.HashTreeProof(
		index,
		&b.Index,
		spec.Wrap(&b.Blob),
		&b.KZGCommitment,
		&b.KZGProof,
		&b.SignedBlockHeader,
		spec.Wrap(&b.KZGCommitmentInclusionProof),
	)
}

func (b *BlobSidecar) IncludeProof(spec *common.Spec, hFn tree.HashFn, beaconBlockBody BeaconBlockBody) error {
	bodyGindex, err := tree.ToGindex64(_blob_kzg_commitments, tree.CoverDepth(_beacon_block_body_length))
	if err != nil {
		return err
	}
	bodyProof := beaconBlockBody.HashTreeProof(spec, hFn, bodyGindex)
	if err != nil {
		return err
	}
	if !bytes.Equal(bodyProof[0][:], b.SignedBlockHeader.Message.BodyRoot[:]) {
		return fmt.Errorf("invalid body root")
	}
	kzgCommitmentGindex, err := tree.ToGindex64(uint64(b.Index), tree.CoverDepth(uint64(spec.MAX_BLOB_COMMITMENTS_PER_BLOCK)))
	if err != nil {
		return err
	}
	kzgCommitmentProof := beaconBlockBody.BlobKZGCommitments.HashTreeProof(spec, hFn, kzgCommitmentGindex)
	if err != nil {
		return err
	}
	b.KZGCommitmentInclusionProof = append(bodyProof[1:], kzgCommitmentProof[1:]...)
	if len(b.KZGCommitmentInclusionProof) != int(spec.KZG_COMMITMENT_INCLUSION_PROOF_DEPTH) {
		return fmt.Errorf("invalid KZG commitment inclusion proof length: %d", len(b.KZGCommitmentInclusionProof))
	}
	return nil
}

func (b *BlobSidecar) VerifyProof(spec *common.Spec, hFn tree.HashFn, beaconBlockBody BeaconBlockBody) error {
	bodyGindex, err := tree.ToGindex64(_blob_kzg_commitments, tree.CoverDepth(_beacon_block_body_length))
	if err != nil {
		return err
	}
	bodyProof := beaconBlockBody.HashTreeProof(spec, hFn, bodyGindex)
	if err != nil {
		return err
	}
	if !bytes.Equal(bodyProof[0][:], b.SignedBlockHeader.Message.BodyRoot[:]) {
		return fmt.Errorf("invalid body root")
	}
	kzgCommitmentGindex, err := tree.ToGindex64(uint64(b.Index), tree.CoverDepth(uint64(spec.MAX_BLOB_COMMITMENTS_PER_BLOCK)))
	if err != nil {
		return err
	}
	kzgCommitmentProof := beaconBlockBody.BlobKZGCommitments.HashTreeProof(spec, hFn, kzgCommitmentGindex)
	if err != nil {
		return err
	}
	if len(b.KZGCommitmentInclusionProof) != int(spec.KZG_COMMITMENT_INCLUSION_PROOF_DEPTH) {
		return fmt.Errorf("invalid KZG commitment inclusion proof length: %d", len(b.KZGCommitmentInclusionProof))
	}
	for i := 0; i < len(b.KZGCommitmentInclusionProof); i++ {
		if !bytes.Equal(b.KZGCommitmentInclusionProof[i][:], kzgCommitmentProof[i][:]) {
			return fmt.Errorf("invalid KZG commitment inclusion proof")
		}
	}
	return nil
}
