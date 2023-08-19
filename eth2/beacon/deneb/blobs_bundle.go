package deneb

import (
	"github.com/protolambda/zrnt/eth2/beacon/common"
	"github.com/protolambda/ztyp/codec"
	"github.com/protolambda/ztyp/tree"
	. "github.com/protolambda/ztyp/view"
)

func BlobsBundleType(spec *common.Spec) *ContainerTypeDef {
	return ContainerType("BlobsBundle", []FieldDef{
		{"commitments", common.KZGCommitmentsType(spec)},
		{"proofs", common.KZGProofsType(spec)},
		{"blobs", BlobsType(spec)},
	})
}

type BlobsBundleView struct {
	*ContainerView
}

func AsBlobsBundle(v View, err error) (*BlobsBundleView, error) {
	c, err := AsContainer(v, err)
	return &BlobsBundleView{c}, err
}

type BlobsBundle struct {
	KZGCommitments common.KZGCommitments `json:"commitments" yaml:"commitments"`
	KZGProofs      common.KZGProofs      `json:"proofs" yaml:"proofs"`
	Blobs          Blobs                 `json:"blobs" yaml:"blobs"`
}

func (bb *BlobsBundle) Deserialize(spec *common.Spec, dr *codec.DecodingReader) error {
	return dr.Container(spec.Wrap(&bb.KZGCommitments), spec.Wrap(&bb.KZGProofs), spec.Wrap(&bb.Blobs))
}

func (bb *BlobsBundle) Serialize(spec *common.Spec, w *codec.EncodingWriter) error {
	return w.Container(spec.Wrap(&bb.KZGCommitments), spec.Wrap(&bb.KZGProofs), spec.Wrap(&bb.Blobs))
}

func (bb *BlobsBundle) ByteLength(spec *common.Spec) uint64 {
	return codec.ContainerLength(spec.Wrap(&bb.KZGCommitments), spec.Wrap(&bb.KZGProofs), spec.Wrap(&bb.Blobs))
}

func (a *BlobsBundle) FixedLength(*common.Spec) uint64 {
	return 0
}

func (bb *BlobsBundle) HashTreeRoot(spec *common.Spec, hFn tree.HashFn) common.Root {
	return hFn.HashTreeRoot(spec.Wrap(&bb.KZGCommitments), spec.Wrap(&bb.KZGProofs), spec.Wrap(&bb.Blobs))
}

func (bb *BlobsBundle) Blinded(spec *common.Spec, hFn tree.HashFn) *BlindedBlobsBundle {
	return &BlindedBlobsBundle{
		KZGCommitments: bb.KZGCommitments,
		KZGProofs:      bb.KZGProofs,
		BlobRoots:      bb.Blobs.Roots(spec, hFn),
	}
}

func (bb *BlobsBundle) GetCommitments() *common.KZGCommitments {
	return &bb.KZGCommitments
}

func (bb *BlobsBundle) GetProofs() *common.KZGProofs {
	return &bb.KZGProofs
}

func (bb *BlobsBundle) GetBlobs() *Blobs {
	return &bb.Blobs
}

func BlindedBlobsBundleType(spec *common.Spec) *ContainerTypeDef {
	return ContainerType("BlindedBlobsBundle", []FieldDef{
		{"commitments", common.KZGCommitmentsType(spec)},
		{"proofs", common.KZGProofsType(spec)},
		{"blob_roots", BlobRootsType(spec)},
	})
}

type BlindedBlobsBundle struct {
	KZGCommitments common.KZGCommitments `json:"commitments" yaml:"commitments"`
	KZGProofs      common.KZGProofs      `json:"proofs" yaml:"proofs"`
	BlobRoots      BlobRoots             `json:"blob_roots" yaml:"blob_roots"`
}

func (b *BlindedBlobsBundle) Deserialize(spec *common.Spec, dr *codec.DecodingReader) error {
	return dr.Container(spec.Wrap(&b.KZGCommitments), spec.Wrap(&b.KZGProofs), spec.Wrap(&b.BlobRoots))
}

func (b *BlindedBlobsBundle) Serialize(spec *common.Spec, w *codec.EncodingWriter) error {
	return w.Container(spec.Wrap(&b.KZGCommitments), spec.Wrap(&b.KZGProofs), spec.Wrap(&b.BlobRoots))
}

func (b *BlindedBlobsBundle) ByteLength(spec *common.Spec) uint64 {
	return codec.ContainerLength(spec.Wrap(&b.KZGCommitments), spec.Wrap(&b.KZGProofs), spec.Wrap(&b.BlobRoots))
}

func (a *BlindedBlobsBundle) FixedLength(*common.Spec) uint64 {
	return 0
}

func (bbb *BlindedBlobsBundle) GetCommitments() *common.KZGCommitments {
	return &bbb.KZGCommitments
}

func (bbb *BlindedBlobsBundle) GetProofs() *common.KZGProofs {
	return &bbb.KZGProofs
}

func (bbb *BlindedBlobsBundle) GetBlobRoots() *BlobRoots {
	return &bbb.BlobRoots
}

func (b *BlindedBlobsBundle) HashTreeRoot(spec *common.Spec, hFn tree.HashFn) common.Root {
	return hFn.HashTreeRoot(spec.Wrap(&b.KZGCommitments), spec.Wrap(&b.KZGProofs), spec.Wrap(&b.BlobRoots))
}
