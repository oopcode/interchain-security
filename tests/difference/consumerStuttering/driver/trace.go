package consumerStuttering

import (
	"encoding/json"
	"io"
	"os"
)

type TraceData struct {
	Meta struct {
		Description       string `json:"description"`
		Format            string `json:"format"`
		FormatDescription string `json:"format-description"`
	} `json:"#meta"`
	States []struct {
		Meta struct {
			Index int `json:"index"`
		} `json:"#meta"`
		ActiveConsumers struct {
			Set []int `json:"#set"`
		} `json:"activeConsumers"`
		AwaitedVscIds struct {
			Set []struct {
				_Tup []int `json:"#tup"`
			} `json:"#set"`
		} `json:"awaitedVSCIds"`
		InitialisingConsumers struct {
			Set []int `json:"#set"`
		} `json:"initialisingConsumers"`
		NextConsumerID int `json:"nextConsumerId"`
		NextVscID      int `json:"nextVSCId"`
		StepCnt        int `json:"stepCnt"`
	} `json:"states"`
	Vars []string `json:"vars"`
}

func LoadTraces(fn string) []TraceData {

	/* #nosec */
	fd, err := os.Open(fn)

	if err != nil {
		panic(err)
	}

	/* #nosec */
	defer fd.Close()

	byteValue, _ := io.ReadAll(fd)

	var ret []TraceData

	err = json.Unmarshal([]byte(byteValue), &ret)

	if err != nil {
		panic(err)
	}

	return ret
}
