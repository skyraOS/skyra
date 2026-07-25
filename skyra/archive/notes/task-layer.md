# task layer

a reality that sits between thread and beings. holds task objects, watches heartbeats, nudges idle beings.

## task object

- issued_by: the being who created the task (owns closure)
- assigned_to: the being doing the work
- impulse: the original request
- last_heartbeat: timestamp, updated every time the assigned being fires through the loop
- resolved: bool, only the issuer can set this to true

## mechanics

- being A tells being B to do something — task layer intercepts and creates a task object
- every time being B fires (thinks, talks to peers, whatever), the task layer sees it as a heartbeat
- if 10 minutes pass with no heartbeat, task layer emits a relation into being B to wake it up
- being B cannot close the task — they report back to being A
- being A says done, task resolves

## what it doesn't do

- no step tracking, no progress — Think's history handles that
- no scheduling or priority — just alive or not
- no task chaining — if that emerges it emerges

## why it's simple

the loop already exists. the task layer just watches traffic and nudges when it stops. the task object is a struct with five fields. the heartbeat is implicit — if a relation passes through with the assigned being's name, the timer resets.
