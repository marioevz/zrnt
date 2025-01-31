package sharding

import (
	"github.com/protolambda/zrnt/eth2/beacon/common"
	"github.com/protolambda/ztyp/codec"
	"github.com/protolambda/ztyp/tree"
	. "github.com/protolambda/ztyp/view"
)

var ShardBlobBodySummaryType = ContainerType("ShardBlobBodySummary", []FieldDef{
	{"commitment", DataCommitmentType},
	{"degree_proof", BLSCommitmentType},
	{"data_root", RootType},
	{"max_priority_fee_per_sample", common.GweiType},
	{"max_fee_per_sample", common.GweiType},
})

type ShardBlobBodySummary struct {
	// The actual data commitment
	Commitment DataCommitment `json:"commitment" yaml:"commitment"`
	// Proof that the degree < commitment.length
	DegreeProof BLSCommitment `json:"degree_proof" yaml:"degree_proof"`
	// Hash-tree-root as summary of the data field
	DataRoot common.Root `json:"data_root" yaml:"data_root"`
	// fee payment fields (EIP 1559 like)
	MaxPriorityFeePerSample common.Gwei `json:"max_priority_fee_per_sample" yaml:"max_priority_fee_per_sample"`
	MaxFeePerSample         common.Gwei `json:"max_fee_per_sample" yaml:"max_fee_per_sample"`
}

func (d *ShardBlobBodySummary) Deserialize(dr *codec.DecodingReader) error {
	return dr.FixedLenContainer(&d.Commitment, &d.DegreeProof, &d.DataRoot, &d.MaxPriorityFeePerSample, &d.MaxFeePerSample)
}

func (d *ShardBlobBodySummary) Serialize(w *codec.EncodingWriter) error {
	return w.FixedLenContainer(&d.Commitment, &d.DegreeProof, &d.DataRoot, &d.MaxPriorityFeePerSample, &d.MaxFeePerSample)
}

func (a *ShardBlobBodySummary) ByteLength() uint64 {
	return ShardBlobBodySummaryType.TypeByteLength()
}

func (a *ShardBlobBodySummary) FixedLength() uint64 {
	return ShardBlobBodySummaryType.TypeByteLength()
}

func (d *ShardBlobBodySummary) HashTreeRoot(hFn tree.HashFn) common.Root {
	return hFn.HashTreeRoot(&d.Commitment, &d.DegreeProof, &d.DataRoot, &d.MaxPriorityFeePerSample, &d.MaxFeePerSample)
}
