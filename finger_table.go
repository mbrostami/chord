package chord

import (
	"math/big"
	"sync"

	"github.com/mbrostami/chord/helpers"
)

// MSIZE is the number of bits in identifier
// in fact only O(log n) are distinct
// ref D - Theorem IV.2
const MSIZE int = helpers.HashSize * 8

type FingerTable struct {
	mutex      sync.RWMutex
	Table      map[int]*RemoteNode // ref D
	TableIndex int                 // to use in fixFinger
	m          int
}

func NewFingerTable() *FingerTable {
	return &FingerTable{
		Table:      make(map[int]*RemoteNode),
		TableIndex: 0,
		m:          MSIZE,
	}
}

func (f *FingerTable) ClosestPrecedingNode(identifier [helpers.HashSize]byte, localNode *Node) *RemoteNode {
	f.mutex.RLock()
	defer f.mutex.RUnlock()
	for m := len(f.Table); m > 0; m-- {
		if f.Table[m] != nil {
			// finger[i] ∈ (n, id)
			if helpers.Between(f.Table[m].Identifier, localNode.Identifier, identifier) {
				return f.Table[m]
			}
		}
	}
	return nil
}

func (f *FingerTable) Set(index int, remoteNode *RemoteNode) {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	f.Table[index] = remoteNode
}

// CalculateIdentifier calculates next identifier
func (f *FingerTable) CalculateIdentifier(localNode *Node) (int, [helpers.HashSize]byte) {
	f.TableIndex++
	if f.TableIndex > f.m {
		f.TableIndex = 1
	}

	meint := new(big.Int)
	meint.SetBytes(localNode.Identifier[:])

	baseint := new(big.Int)
	baseint.SetUint64(2)

	powint := new(big.Int)
	powint.SetInt64(int64(f.TableIndex - 1))

	var biggest [helpers.HashSize + 1]byte
	for i := range biggest {
		biggest[i] = 255
	}

	tmp := new(big.Int)
	tmp.SetInt64(1)

	modint := new(big.Int)
	modint.SetBytes(biggest[:helpers.HashSize])
	modint.Add(modint, tmp)

	target := new(big.Int)
	target.Exp(baseint, powint, modint)
	target.Add(meint, target)
	target.Mod(target, modint)

	bytes := target.Bytes()
	diff := helpers.HashSize - len(bytes)
	if diff > 0 {
		tmp := make([]byte, helpers.HashSize)
		//pad with zeros
		for i := 0; i < diff; i++ {
			tmp[i] = 0
		}
		for i := diff; i < helpers.HashSize; i++ {
			tmp[i] = bytes[i-diff]
		}
		bytes = tmp
	}
	var identifier [helpers.HashSize]byte
	copy(identifier[:helpers.HashSize], bytes[:helpers.HashSize])
	return f.TableIndex, identifier
}
