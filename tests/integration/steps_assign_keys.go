package main

// submits a consumer-removal proposal and removes the chain
func stepsAssignConsumerKey(consumerName string) []Step {
	s := []Step{
		{
			action: assignConsumerPubKeyAction{
				chain:          chainID("consu"),
				validator:      validatorID("alice"),
				consumerPubKey: `{"type":"tendermint/PubKeyEd25519","value":"GJuUXISPjcWRIbEdzLTtVHzhnt9T98URH/gOA8KB7fA="}`,
			},
			state: State{
				chainID("consu"): ChainState{
					AssignedKeys: &map[validatorID]string{
						validatorID("alice"): "GJuUXISPjcWRIbEdzLTtVHzhnt9T98URH/gOA8KB7fA=",
					},
				},
			},
		},
	}

	return s
}
