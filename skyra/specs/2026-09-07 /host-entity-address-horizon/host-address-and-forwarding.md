# Host address and forwarding

**Date:** 2026-09-07  
**Status:** Forward traversal accepted; forwarding mechanics remain open.

The lookup path is Entity → EntityAddress → HostAddress → Host. Entity holds multiple EntityAddress
bindings. Each EntityAddress owns its Representation method and can be reached through multiple
HostAddresses. Each HostAddress holds contact information and points to a Host identity.

EntityAddress identity and its representation contract are stable; the host routes through which it
can be reached may change. The whole struct is not immutable if its HostAddresses collection changes.
Hosts sharing an EntityAddress must fulfill the same endpoint contract, not merely run the same software.
Route selection and updates remain to be defined.

Host now holds only its ID. The earlier Host.SelfAddress, EntityAddress.Entity, and EntityAddress.Host
pointers are removed from the proposal. No reverse lookup or circular traversal is required by this
shape. Missing-pointer meaning, route selection, and forwarding remain to be defined.

Egress must distinguish:

- This CTU's endpoint: local delivery.
- Another endpoint on the same host: route to that entity; transport may still be required.
- An endpoint on another host: route across hosts.

A matching HostID identifies the same hosting environment, not necessarily the same CTU or physical
machine. A different address identifies another endpoint, not necessarily another device. This matters
when a host represents an account/environment containing multiple execution resources.

The CTU's own endpoint representation and address-matching rules remain open. The accepted stub is in
[the structs/methods proposal](proposal-entity-address-host.md).
