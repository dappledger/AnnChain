package state

import (
	"sort"

	cvtypes "github.com/dappledger/AnnChain/src/types"
)

type ModifiedUndoItfc interface {
	Undo(state cvtypes.StateItfc)
	Copy() ModifiedUndoItfc
}

type StVersion struct {
	id      int
	journey int
}

type RevertLayer struct {
	state   cvtypes.StateItfc
	journey []ModifiedUndoItfc

	versions  []StVersion // index:versionIndex, value:index of journey
	versionID int
}

func (rl *RevertLayer) Init(state cvtypes.StateItfc) {
	rl.state = state
	rl.journey = make([]ModifiedUndoItfc, 0)
	rl.versions = make([]StVersion, 0)
	rl.versionID = 0
}

func (rl *RevertLayer) Copy() *RevertLayer {
	cprl := &RevertLayer{}
	cprl.state = rl.state
	cprl.journey = make([]ModifiedUndoItfc, 0, len(rl.journey))
	for i := range rl.journey {
		cprl.journey = append(cprl.journey, rl.journey[i].Copy())
	}
	cprl.versions = make([]StVersion, len(rl.versions))
	copy(cprl.versions, rl.versions)
	return cprl
}

func (rl *RevertLayer) Snapshot() int {
	curv := rl.versionID
	rl.versions = append(rl.versions, StVersion{id: curv, journey: len(rl.journey)})
	rl.versionID++
	return curv
}

func (rl *RevertLayer) RevertToVerstion(v int) {
	idx := sort.Search(len(rl.versions), func(i int) bool {
		return rl.versions[i].id >= v
	})
	if idx == len(rl.versions) || rl.versions[idx].id != v {
		return
	}
	if len(rl.journey) == 0 {
		rl.versions = rl.versions[:idx]
		return
	}
	snapshot := rl.versions[idx].journey

	for i := len(rl.journey) - 1; i >= snapshot; i-- {
		rl.journey[i].Undo(rl.state)
	}
	rl.journey = rl.journey[:snapshot]
	rl.versions = rl.versions[:idx]
}

func (rl *RevertLayer) addJourney(r ModifiedUndoItfc) {
	rl.journey = append(rl.journey, r)
}
