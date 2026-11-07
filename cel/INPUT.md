# Input given to each of the 7 sub agents

Each agent is one `claude -p` process, run exactly as
`controller/apprentice/typesys-prototype/drive/claude.go` runs an ask
(`ClaudeModel.arguments`): model claude-opus-5-5, effort high, `--tools ""`
(no tools at all), no MCP, no hooks, and the act-frame system prompt
(systemHead + actInstructions + systemTail) as `--system-prompt`.
Its stdout (`--output-format json`) is redirected straight into `agent-N.json`, unedited.

Parameters: Goal $1; Tools $2–$4 (3 tools); Context $5–$6 (passed context).

## System prompt

See `system-prompt.txt` in this directory: `drive.System(tree.KindActs)` printed from the Go, with the typed return: `return <0|1> $out [$err]` builds `{exit: exit status, stdout: bytes, stderr?: bytes}`; the null rule spelled out; instructions made only of parameters.

## The ask (the prompt, identical for all 7)

```
Goal:
$1 (string): make sure the service is healthy, and fix it if it isn't

Tools:
$2: curl
$3: systemctl
$4: journalctl

Context:
$5 (string):
HTTP/1.1 503 Service Unavailable
retry-after: 30
$6 ({"restarts":number,"unit":string,"url":string}): {"restarts":3,"unit":"api.service","url":"http://localhost:8080/health"}
```
