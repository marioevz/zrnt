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

func (bb *BlobsBundle) HashTreeProof(spec *common.Spec, hFn tree.HashFn, index tree.Gindex) []common.Root {
	return hFn.HashTreeProof(index, spec.Wrap(&bb.KZGCommitments), spec.Wrap(&bb.KZGProofs), spec.Wrap(&bb.Blobs))
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
