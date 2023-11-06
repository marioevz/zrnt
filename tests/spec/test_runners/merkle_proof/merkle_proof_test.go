package merkle_proof

import (
	"bytes"
	"encoding/hex"
	"io"
	"testing"

	"github.com/golang/snappy"
	"github.com/protolambda/zrnt/eth2/beacon/common"
	"github.com/protolambda/zrnt/eth2/beacon/deneb"
	"github.com/protolambda/zrnt/eth2/beacon/phase0"
	"github.com/protolambda/zrnt/eth2/configs"
	"github.com/protolambda/zrnt/tests/spec/test_util"
	"github.com/protolambda/ztyp/codec"
	"github.com/protolambda/ztyp/tree"
	"gopkg.in/yaml.v2"
)

type MerkleProofTestCase interface {
	Run(t *testing.T)
}

type ObjAllocator func() interface{}

type MerkleTestType string

const (
	SingleMerkleProof MerkleTestType = "single_merkle_proof"
)

var AllMerkleTestTypes = []MerkleTestType{SingleMerkleProof}

func testCaseByTypeName(name MerkleTestType, spec *common.Spec, typeName string, obj interface{}, readPart test_util.TestPartReader) MerkleProofTestCase {
	switch name {
	case "single_merkle_proof":
		return &SingleMerkleProofTestCase{
			TypeName:       typeName,
			Spec:           spec,
			Object:         obj,
			TestPartReader: readPart,
		}
	default:
		panic("unrecognized merkle proof test type: " + name)
	}
}

type SingleMerkleProofTestCase struct {
	TypeName       string
	Spec           *common.Spec
	Object         interface{}
	TestPartReader test_util.TestPartReader

	Leaf      tree.Root
	LeafIndex tree.Gindex
	Branch    []tree.Root
}

type ProofYAML struct {
	Leaf      string   `yaml:"leaf"`
	LeafIndex uint64   `yaml:"leaf_index"`
	Branch    []string `yaml:"branch"`
}

func (c *SingleMerkleProofTestCase) Run(t *testing.T) {
	{
		p := c.TestPartReader.Part("proof.yaml")
		dec := yaml.NewDecoder(p)
		proof := &ProofYAML{}
		test_util.Check(t, dec.Decode(proof))
		test_util.Check(t, p.Close())
		{
			leaf, err := hex.DecodeString(proof.Leaf[2:])
			test_util.Check(t, err)
			copy(c.Leaf[:], leaf)
			c.LeafIndex = tree.Gindex64(proof.LeafIndex)
			for _, b := range proof.Branch {
				branch, err := hex.DecodeString(b[2:])
				test_util.Check(t, err)
				var root tree.Root
				copy(root[:], branch)
				c.Branch = append(c.Branch, root)
			}
		}
		var (
			objRoot   tree.Root
			calcProof []tree.Root
		)
		if specObj, ok := c.Object.(common.SpecObj); ok {
			objRoot = specObj.HashTreeRoot(c.Spec, tree.GetHashFn())
			calcProof = specObj.HashTreeProof(c.Spec, tree.GetHashFn(), c.LeafIndex)
		} else if htpObj, ok := c.Object.(tree.HTP); ok {
			objRoot = htpObj.HashTreeRoot(tree.GetHashFn())
			calcProof = htpObj.HashTreeProof(tree.GetHashFn(), c.LeafIndex)
		}
		if len(calcProof) != len(c.Branch) {
			t.Fatalf("expected %d proof elements, got %d", len(c.Branch), len(calcProof))
		}

		for i := range calcProof {
			if calcProof[i] != c.Branch[i] {
				t.Errorf("proof element %d is incorrect, expected %s, got %s", i, c.Branch[i].String(), calcProof[i].String())
			}
		}

		if !tree.VerifyProof(tree.GetHashFn(), c.Branch, c.LeafIndex, objRoot, c.Leaf) {
			t.Error("proof verification failed")
		}
	}
}

var objs = map[test_util.ForkName]map[MerkleTestType]map[string]ObjAllocator{
	"phase0":    {},
	"altair":    {},
	"bellatrix": {},
	"capella":   {},
	"deneb":     {},
}

func init() {
	base := map[string]ObjAllocator{
		"BeaconBlockBody": func() interface{} { return new(phase0.BeaconBlockBody) },
	}
	for _, testType := range AllMerkleTestTypes {
		objs["phase0"][testType] = make(map[string]ObjAllocator)
		objs["altair"][testType] = make(map[string]ObjAllocator)
		objs["bellatrix"][testType] = make(map[string]ObjAllocator)
		objs["capella"][testType] = make(map[string]ObjAllocator)
		objs["deneb"][testType] = make(map[string]ObjAllocator)

		for k, v := range base {
			// objs["phase0"][k] = v
			// objs["altair"][k] = v
			// objs["bellatrix"][k] = v
			// objs["capella"][k] = v
			objs["deneb"][testType][k] = v
		}
		// objs["altair"]["BeaconBlockBody"] = func() interface{} { return new(altair.BeaconBlockBody) }
		// objs["bellatrix"]["BeaconBlockBody"] = func() interface{} { return new(altair.BeaconBlockBody) }
		// objs["capella"]["BeaconBlockBody"] = func() interface{} { return new(altair.BeaconBlockBody) }
		objs["deneb"][testType]["BeaconBlockBody"] = func() interface{} { return new(deneb.BeaconBlockBody) }
	}

}

func runMerkleProofTest(fork test_util.ForkName, testType MerkleTestType, name string, alloc ObjAllocator, spec *common.Spec) func(t *testing.T) {
	return func(t *testing.T) {
		path := "merkle_proof/" + string(testType) + "/" + name
		test_util.RunHandler(t, path, func(t *testing.T, forkName test_util.ForkName, readPart test_util.TestPartReader) {
			// Allocate an empty value to decode into later for testing.
			obj := alloc()
			spec := readPart.Spec()
			{
				p := readPart.Part("object.ssz_snappy")
				data, err := io.ReadAll(p)
				test_util.Check(t, err)
				uncompressed, err := snappy.Decode(nil, data)

				r := bytes.NewReader(uncompressed)
				if obj, ok := obj.(common.SpecObj); ok {

					if err := obj.Deserialize(spec, codec.NewDecodingReader(r, uint64(len(uncompressed)))); err != nil {
						t.Fatal(err)
					}
				}
				test_util.Check(t, err)
				test_util.Check(t, p.Close())
				test_util.Check(t, err)
			}

			c := testCaseByTypeName(testType, spec, name, obj, readPart)
			c.Run(t)
		}, spec, fork)
	}
}
func TestMerkeProof(t *testing.T) {
	t.Parallel()
	for configName, config := range map[string]*common.Spec{
		"minimal": configs.Minimal,
		"mainnet": configs.Mainnet,
	} {
		config := config
		t.Run(configName, func(t *testing.T) {
			for fork, objByName := range objs {
				t.Run(string(fork), func(t *testing.T) {
					for testType, v := range objByName {
						for k, v := range v {
							t.Run(k, runMerkleProofTest(fork, testType, k, v, config))
						}
					}
				})
			}
		})
	}
}
