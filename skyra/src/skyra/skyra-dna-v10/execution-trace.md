# v10 execution trace — actual passing values

**Date:** 2026-08-24  
**Source:** `go test -run TestAcquireEndToEnd -count=1 -v ./src/skyra/skyra-dna-v10/acquire`

The lines below are the instrumented output from the passing run, grouped only with annotations. Values
are printed at the boundary where they return; this is not a reconstructed `%v` dump.

## 1. Read-only discover returns a proposal

```text
ASSERT start acquire horizon.grants=0 kernel.mints=0 (read-only)
SEARCH input="weather in San Francisco" -> names=[weather.example]
CHECK block=ENTITY/STAMPED+CHECKED stamped_by=search contract=names result=PASS
RESOLVE input={names:[weather.example] hints:{service:https triple:inet/stream/tcp}} -> discover RegistrationProposal{candidates:[{name:weather.example endpoint:203.0.113.7:443 novelty:{known:false history:"no registry history"}}]}
CHECK block=ENTITY/STAMPED+CHECKED stamped_by=resolve contract=proposal result=PASS
```

The live resolve consumed explicit hints. At proposal return, no grant is held and kernel Mint has not
run.

## 2. Novel first-contact faults, parks, and re-drives

```text
FAULT target=skyra://privilege/user endpoint=203.0.113.7:443 kind=Poke scope={purpose:"weather in San Francisco" action:Poke}
PRIVILEGE DECIDE kind=Poke novelty={known:false history:"no registry history"} policy=Ask
PARK move=reach site=Cut{target:skyra://privilege/user} in_flight=permission.1 writes=grant.1 grant_landed=false
ASSERT after issuing Drive returned outcome=NOTYET result=nil in_flight=1 kernel.mints=0 endpoint_probe_calls=0 privilege_peer_calls=0
HUMAN APPROVE id=permission.1 kind=Poke (separate event stack)
RECOGNIZE id=permission.1 landed=grant.1 grant={kind:Poke subject:203.0.113.7:443} -> RE-DRIVE
```

The assertion executes only after `Drive` returned, and approval is invoked afterward. Thus the grant
cannot have landed on the issuing stack. The test deliberately seeds a peer under the privilege
spelling; its zero call count proves the fault crossed the trap instead of peer dispatch. The endpoint
probe's zero count proves first contact did not occur before the Poke grant.

## 3. Candidate hand-off and all four safe probe returns

```text
HANDOFF discover Candidate{endpoint:203.0.113.7:443} -> test input
PROBE reach input=203.0.113.7:443 -> Reachability{up:true why:"mock connect answered"}
CHECK block=ENTITY/STAMPED+CHECKED stamped_by=endpoint:dns:203.0.113.7:443/tcp/stream contract=reach result=PASS
PROBE handshake input=203.0.113.7:443 -> {transport:{port:443 tls:true alpn:h2} identity:{cert_host:"weather.example"}}
CHECK block=ENTITY/STAMPED+CHECKED stamped_by=endpoint:dns:203.0.113.7:443/tcp/stream contract=handshake result=PASS
PROBE challenge input=safe-unauthenticated-GET -> Auth{method:bearer cred_type:oauth2-token verified:false}
CHECK block=ENTITY/STAMPED+CHECKED stamped_by=endpoint:dns:203.0.113.7:443/tcp/stream contract=auth result=PASS
PROBE describe input=safe-OPTIONS -> Fitness{operation:"GET /forecast" request:ForecastRequest response:ForecastReply required:[forecast temperature_c]}
CHECK block=ENTITY/STAMPED+CHECKED stamped_by=endpoint:dns:203.0.113.7:443/tcp/stream contract=fitness result=PASS
```

Challenge deliberately discovers a credential type without obtaining or presenting one. The mock's
unsafe-call counter remains zero.

## 4. Join, contract hand-off, admission, and kernel mint

```text
FITNESS MATCH intent="weather in San Francisco" operation="GET /forecast" -> selected=true
TEST JOIN fragments -> ConnectionContract{transport:{port:443 tls:true alpn:h2} identity:{cert_host:"weather.example"} auth:{method:bearer cred_type:oauth2-token verified:false} fitness:{operation:"GET /forecast" response:ForecastReply required:[forecast temperature_c]} partial:true}
HANDOFF ConnectionContract -> admit input={endpoint:203.0.113.7:443 operation:"GET /forecast"}
ADMIT PACKAGE input={endpoint:203.0.113.7:443 contract.operation:"GET /forecast"} -> RegistrationRequest{endpoint:203.0.113.7:443 block:ENTITY/STAMPED+CHECKED contract.operation:"GET /forecast"}
FAULT target=skyra://privilege/user endpoint=203.0.113.7:443 kind=Register scope={purpose:"weather in San Francisco" action:Register}
PRIVILEGE DECIDE kind=Register novelty={known:false history:"no registry history"} policy=Ask
PARK move=admit site=Cut{target:skyra://privilege/user} in_flight=permission.2 writes=grant.2 grant_landed=false
HUMAN APPROVE id=permission.2 kind=Register (separate event stack)
PRIVILEGE TRAP decision=allow-register -> kernel (privilege holds no Mint)
KERNEL MINT endpoint=203.0.113.7:443 block=ENTITY/STAMPED+CHECKED contract={operation:GET /forecast response:ForecastReply required:[forecast temperature_c]}
RECOGNIZE id=permission.2 landed=grant.2 grant={kind:Register subject:203.0.113.7:443} -> RE-DRIVE
ADMIT -> Capability{name:endpoint:dns:203.0.113.7:443/tcp/stream endpoint:203.0.113.7:443 block:ENTITY/STAMPED+CHECKED contract.operation:"GET /forecast"}
REGISTRY READBACK endpoint=203.0.113.7:443 block=ENTITY/STAMPED+CHECKED contract.response={name:ForecastReply required:[forecast temperature_c]}
QUIESCENCE retired=9/9 in_flight=0 result=Capability
```

Privilege prints the decision and trap; kernel prints the Mint. The capability returned by admit is the
same endpoint/contract read back from the Entity registry.

## 5. The registered output contract is the dispatch gate

```text
CHECK block=ENTITY/STAMPED+CHECKED stamped_by=endpoint:dns:203.0.113.7:443/tcp/stream contract=ForecastReply result=PASS
USE ENTITY honoring reply=map[forecast:sunny temperature_c:19] -> SUCCESS
CHECK block=ENTITY/STAMPED+CHECKED stamped_by=endpoint:dns:203.0.113.7:443/tcp/stream contract=ForecastReply result=FAIL value=map[forecast:sunny]
USE ENTITY violating reply=map[forecast:sunny] -> FAILURE
USE HOST same violating reply=map[forecast:sunny] -> SUCCESS (BELIEVED, no contract gate)
```

The missing `temperature_c` is rejected only on the acquired Entity path. `stamped_by` is the actual
`Observation.Segment.By` value set before the check, not a descriptive label. The same reply from a
Host is believed, proving the difference is the block trust rule rather than the fake's return status.

## 6. Reacquire takes the known fast path

```text
RESOLVE input={names:[weather.example] hints:{service:https triple:inet/stream/tcp}} -> discover RegistrationProposal{candidates:[{name:weather.example endpoint:203.0.113.7:443 novelty:{known:true history:"registered entity capability"}}]}
HANDOFF discover Candidate{endpoint:203.0.113.7:443} -> test input
PROBE reach input=203.0.113.7:443 -> Reachability{up:true why:"mock connect answered"}
PROBE handshake input=203.0.113.7:443 -> {transport:{port:443 tls:true alpn:h2} identity:{cert_host:"weather.example"}}
PROBE challenge input=safe-unauthenticated-GET -> Auth{method:bearer cred_type:oauth2-token verified:false}
PROBE describe input=safe-OPTIONS -> Fitness{operation:"GET /forecast" request:ForecastRequest response:ForecastReply required:[forecast temperature_c]}
FITNESS MATCH intent="weather in San Francisco" operation="GET /forecast" -> selected=true
TEST JOIN fragments -> ConnectionContract{transport:{port:443 tls:true alpn:h2} identity:{cert_host:"weather.example"} auth:{method:bearer cred_type:oauth2-token verified:false} fitness:{operation:"GET /forecast" response:ForecastReply required:[forecast temperature_c]} partial:true}
HANDOFF ConnectionContract -> admit input={endpoint:203.0.113.7:443 operation:"GET /forecast"}
ADMIT PACKAGE input={endpoint:203.0.113.7:443 contract.operation:"GET /forecast"} -> RegistrationRequest{endpoint:203.0.113.7:443 block:ENTITY/STAMPED+CHECKED contract.operation:"GET /forecast"}
REACQUIRE known=true -> SUCCESS faults=0 grants=0 wakes=0 new_mints=0 retired=9/9
```

The counters are deltas captured immediately before reacquire. The known endpoint resolves directly to
the registered Entity for test, and admit returns the existing capability.
