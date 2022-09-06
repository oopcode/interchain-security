package consumerStuttering

import (
	"encoding/json"
	"io"
	"os"
)

type RawState struct {
	Meta struct {
		Index int `json:"index"`
	} `json:"#meta"`
	ActionKind      string `json:"actionKind"`
	ActiveConsumers struct {
		Set []int `json:"#set"`
	} `json:"activeConsumers"`
	AwaitedVscIds struct {
		Set []struct {
			Tup []int `json:"#tup"`
		} `json:"#set"`
	} `json:"awaitedVSCIds"`
	InitialisingConsumers struct {
		Set []int `json:"#set"`
	} `json:"initialisingConsumers"`
	NextConsumerID int `json:"nextConsumerId"`
	NextVscID      int `json:"nextVSCId"`
}

type RawTraceData struct {
	Meta struct {
		Description       string `json:"description"`
		Format            string `json:"format"`
		FormatDescription string `json:"format-description"`
	} `json:"#meta"`
	States []RawState `json:"states"`
	Vars   []string   `json:"vars"`
}

type State struct {
	Kind                  string
	InitialisingConsumers []int
	ActiveConsumers       []int
	AwaitedVscIds         [][]int
}

type TraceData struct {
	States []State
}

func convertState(in RawState) (out State) {
	out.InitialisingConsumers = in.InitialisingConsumers.Set
	out.ActiveConsumers = in.ActiveConsumers.Set
	out.AwaitedVscIds = [][]int{}
	out.Kind = in.ActionKind
	for _, pair := range in.AwaitedVscIds.Set {
		out.AwaitedVscIds = append(out.AwaitedVscIds, pair.Tup)
	}
	return out
}

func convert(in RawTraceData) (out TraceData) {
	out.States = []State{}
	for _, state := range in.States {
		out.States = append(out.States, convertState(state))
	}
	return out
}

func LoadTraces(fn string) []TraceData {

	data := loadRawTraces(fn)

	ret := []TraceData{}

	for _, trace := range data {
		ret = append(ret, convert(trace))
	}

	return ret
}

func loadRawTraces(fn string) []RawTraceData {

	/* #nosec */
	fd, err := os.Open(fn)

	if err != nil {
		panic(err)
	}

	/* #nosec */
	defer fd.Close()

	byteValue, _ := io.ReadAll(fd)

	var ret []RawTraceData

	err = json.Unmarshal([]byte(byteValue), &ret)

	if err != nil {
		panic(err)
	}

	return ret
}

// getDifferentInt finds an int in a that is not in b
func getDifferentInt(a, b []int) *int {
	exist := map[int]bool{}
	for _, x := range b {
		exist[x] = true
	}
	for _, x := range a {
		if !exist[x] {
			return &x
		}
	}
	return nil
}

// getDifferentIntPair finds an int pair in a that is not in b
func getDifferentIntPair(a, b [][]int) []int {
	// TODO: improve perf
	for _, x := range a {
		seen := false
		for _, y := range b {
			if y[0] == x[0] && y[1] == x[1] {
				seen = true
			}
		}
		if !seen {
			return x
		}
	}
	return nil
}
