# routing rules

from v.04 derive-present-spec. not yet implemented in v.05.

## self-reference drop

an entity cannot target itself.

## auto-close on return

if the entity is returning to whoever called it, close the detour exchange.

## ~ref departure close

if the entity is addressing a new peer with a ~ref, close the current exchange.

## parent block

if the target is the entity's parent in an active exchange the entity didn't open, block the message.

## departure visibility

when an entity opens a new exchange while a parent exchange is waiting, annotate the parent exchange. when the entity returns, annotate again. these are world-level annotations — not messages from the entity, just visible bookkeeping in the exchange history.

## relationship enforcement

the world checks the sender's relationship list before routing. if the target is not in the sender's relationships, the message is dropped.
