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
	KZGCommitments common.KZGCommitments `json:"kzg_commitments" yaml:"kzg_commitments"`
	KZGProofs      common.KZGProofs      `json:"kzg_proofs" yaml:"kzg_proofs"`
	Blobs          Blobs                 `json:"blobs" yaml:"blobs"`
}

func (s *BlobsBundle) Deserialize(spec *common.Spec, dr *codec.DecodingReader) error {
	return dr.Container(spec.Wrap(&s.KZGCommitments), spec.Wrap(&s.KZGProofs), spec.Wrap(&s.Blobs))
}

func (s *BlobsBundle) Serialize(spec *common.Spec, w *codec.EncodingWriter) error {
	return w.Container(spec.Wrap(&s.KZGCommitments), spec.Wrap(&s.KZGProofs), spec.Wrap(&s.Blobs))
}

func (s *BlobsBundle) ByteLength(spec *common.Spec) uint64 {
	return codec.ContainerLength(spec.Wrap(&s.KZGCommitments), spec.Wrap(&s.KZGProofs), spec.Wrap(&s.Blobs))
}

func (a *BlobsBundle) FixedLength(*common.Spec) uint64 {
	return 0
}

func (s *BlobsBundle) HashTreeRoot(spec *common.Spec, hFn tree.HashFn) common.Root {
	return hFn.HashTreeRoot(spec.Wrap(&s.KZGCommitments), spec.Wrap(&s.KZGProofs), spec.Wrap(&s.Blobs))
}

func BlindedBlobsBundleType(spec *common.Spec) *ContainerTypeDef {
	return ContainerType("BlindedBlobsBundle", []FieldDef{
		{"commitments", common.KZGCommitmentsType(spec)},
		{"proofs", common.KZGProofsType(spec)},
		{"blob_roots", BlobRootsType(spec)},
	})
}

type BlindedBlobsBundle struct {
	KZGCommitments common.KZGCommitments
	KZGProofs      common.KZGProofs
	BlobRoots      BlobRoots
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

func (b *BlindedBlobsBundle) HashTreeRoot(spec *common.Spec, hFn tree.HashFn) common.Root {
	return hFn.HashTreeRoot(spec.Wrap(&b.KZGCommitments), spec.Wrap(&b.KZGProofs), spec.Wrap(&b.BlobRoots))
}
