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

type TraceData struct {
	Meta struct {
		Description       string `json:"description"`
		Format            string `json:"format"`
		FormatDescription string `json:"format-description"`
	} `json:"#meta"`
	States []State  `json:"states"`
	Vars   []string `json:"vars"`
}

func convertState(in RawState) (out State) {
	//TODO:
	return
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
