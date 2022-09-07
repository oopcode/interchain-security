package consumerStuttering

import (
	"testing"
)

// checkNoUnbondEarly checks that for all vscIds which are still awaited, the refCnts
// for all unbonding operations associated to the vscId are positive
func checkNoUnbondEarly(t *testing.T, refCnt map[uint64]int, vscIdToOpids map[uint64][]uint64,
	awaitedVscIds [][]int) {
	for _, pair := range awaitedVscIds {
		vscId := pair[1]
		for _, opId := range vscIdToOpids[uint64(vscId)] {
			if refCnt[opId] < 1 {
				t.Fatalf("fail checkNoUnbondEarly")
			}
		}
	}
}

// checkNoUnbondLate checks that for all vscId < valUpdateId: if there is NOT an awaited
// maturity for that vscID, then the refCnts for all unbonding operations associated to
// the vscID are 0
func checkNoUnbondLate(t *testing.T,
	refCnt map[uint64]int,
	vscIdToOpids map[uint64][]uint64,
	awaitedVscIds [][]int,
	maxVscIdToCheck uint64) {

	stillAwaiting := make([]bool, maxVscIdToCheck+2)
	for _, pair := range awaitedVscIds {
		vscId := pair[1]
		if uint64(vscId) <= maxVscIdToCheck {
			stillAwaiting[vscId] = true
		}
	}
	for vscId := uint64(0); vscId <= maxVscIdToCheck; vscId++ {
		if !stillAwaiting[vscId] {
			for _, opId := range vscIdToOpids[uint64(vscId)] {
				if 0 < refCnt[opId] {
					t.Fatalf("fail checkNoUnbondLate")
				}
			}
		}
	}
}
