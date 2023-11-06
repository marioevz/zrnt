package common

import (
	"github.com/protolambda/ztyp/codec"
	"github.com/protolambda/ztyp/tree"
	. "github.com/protolambda/ztyp/view"
)

type CommitteeIndices []ValidatorIndex

func (p *CommitteeIndices) Deserialize(spec *Spec, dr *codec.DecodingReader) error {
	return dr.List(func() codec.Deserializable {
		i := len(*p)
		*p = append(*p, ValidatorIndex(0))
		return &((*p)[i])
	}, ValidatorIndexType.TypeByteLength(), uint64(spec.MAX_VALIDATORS_PER_COMMITTEE))
}

func (a CommitteeIndices) Serialize(_ *Spec, w *codec.EncodingWriter) error {
	return w.List(func(i uint64) codec.Serializable {
		return a[i]
	}, ValidatorIndexType.TypeByteLength(), uint64(len(a)))
}

func (a CommitteeIndices) ByteLength(*Spec) uint64 {
	return ValidatorIndexType.TypeByteLength() * uint64(len(a))
}

func (*CommitteeIndices) FixedLength(*Spec) uint64 {
	return 0
}

func (p CommitteeIndices) HashTreeRoot(spec *Spec, hFn tree.HashFn) Root {
	return hFn.Uint64ListHTR(func(i uint64) uint64 {
		return uint64(p[i])
	}, uint64(len(p)), uint64(spec.MAX_VALIDATORS_PER_COMMITTEE))
}

func (p CommitteeIndices) HashTreeProof(spec *Spec, hFn tree.HashFn, index tree.Gindex) []Root {
	return hFn.Uint64ListHTP(func(i uint64) uint64 {
		return uint64(p[i])
	}, uint64(len(p)), uint64(spec.MAX_PROPOSER_SLASHINGS), index)
}

func (c *Phase0Preset) CommitteeIndices() ListTypeDef {
	return ListType(ValidatorIndexType, uint64(c.MAX_VALIDATORS_PER_COMMITTEE))
}
