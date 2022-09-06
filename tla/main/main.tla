---- MODULE main ----

EXTENDS Integers, FiniteSets, Sequences, TLC, Apalache

VARIABLES
    \* @type: Int;
    nextVSCId,
    \* @type: Int;
    nextConsumerId,
    \* @type: Set(Int);
    initialisingConsumers,
    \* @type: Set(Int); 
    activeConsumers,
    \* @type: Set(Int);
    newUnbondings,

    (*
    Need to get vsc on a channel and map it to unbondings
    
    *)
    \* Need to get VSC on 

InitConsumer ==
    /\ initialisingConsumers' = initialisingConsumers \cup {nextConsumerId}
    /\ nextConsumerId' = nextConsumerId + 1
    
ActivateConsumer == 
    \E c \in initialisingConsumers:
        /\ initialisingConsumers' = initialisingConsumers \ {c}
        /\ activeConsumers' = activeConsumers \cup {c}

StopConsumer == 
    \/ \E c \in initialisingConsumers:
        /\ initialisingConsumers' = initialisingConsumers \ {c}
        /\ UNCHANGED activeConsumers
    \/ \E c \in activeConsumers:
        /\ activeConsumers' = activeConsumers \ {c}
        /\ UNCHANGED initialisingConsumers

TrackNewUnbonding == 
    /\ newUnbondings' = newUnbondings \cup {nextUnbondingId}
    /\ nextUnbondingId' = nextUnbondingId + 1

EndBlock == 
    \* CompleteMatured
    \* SendValidatorUpdates
    
RecvMaturity == 

Init == 

Next == 
    \/ InitConsumer
    \/ ActivateConsumer 
    \/ StopConsumer
    \/ TrackNewUnbonding
    \/ RecvMaturity
    \/ EndBlock

Inv == Len(x) < 100


====
