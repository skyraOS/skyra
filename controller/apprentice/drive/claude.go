package drive

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/anthropics/anthropic-sdk-go/packages/param"
	"github.com/anthropics/anthropic-sdk-go/shared/constant"

	"skyra/apprentice/tree"
)

// ClaudeModel is the model the machine startles: Claude, asked once per
// AskModel. Within the ask it may expand truncated values (ExpandLimit
// times, never counted against the refinement bound) and ends by writing
// its guard.
type ClaudeModel struct {
	Messages messenger // the Messages API; NewClaudeModel's client by default
	Model    string
	Effort   anthropic.BetaOutputConfigEffort
}

type messenger interface {
	New(ctx context.Context, p anthropic.BetaMessageNewParams, opts ...option.RequestOption) (*anthropic.BetaMessage, error)
}

// NewClaudeModel is Claude Opus 5.5 through the default client, which reads
// its credentials from the environment.
func NewClaudeModel() *ClaudeModel {
	c := anthropic.NewClient()
	return &ClaudeModel{Messages: &c.Beta.Messages, Model: "claude-opus-5-5", Effort: anthropic.BetaOutputConfigEffortHigh}
}

// The system prompt is fixed, so it caches.
const claudeSystem = `You write guards for a tree that learns. Each time it reaches a place it has not learned, it asks you what to do there. What you write is stored and reused: the next time anything arrives at the same base with values your guard selects, it runs with no model call.

Observations are named $1, $2, ... in the frame's scope. A tool observation names a binary the frame may run. Values are truncated; call expand to read more of one.

A guard is CEL that returns an instruction or null:
  size($2) > 0 ? "$1 $2" : null
  $3.exit != 0 ? "return $3" : $3.contains("b.log") ? "summarize $1 $3" : null
Conditions read observations by $n and their fields by key ($3.size, $3["content-type"]). Only $n listed in the base are learned; a guard reading another runs once and is not stored.

An instruction is read by its first field:
  return $n            returns $n to the frame's caller
  $t args...           $t is a tool: runs its binary with the args, each $n replaced by its value
  task $a $b ...       delegates to a child frame named task, passing only $n; the child must be passed fewer tools than this frame holds

Your guard must select something for what has arrived. When you are done, call guard with it.`

func (m *ClaudeModel) tools() []anthropic.BetaToolUnionParam {
	expand := anthropic.BetaToolParam{
		Name:        "expand",
		Description: param.NewOpt("Read more of a truncated value: up to the show limit of $n's text from the byte offset from."),
		Strict:      param.NewOpt(true),
		InputSchema: anthropic.BetaToolInputSchemaParam{
			Properties: map[string]any{
				"n":    map[string]any{"type": "integer", "description": "the observation's number, as in $n"},
				"from": map[string]any{"type": "integer", "description": "the byte offset to read from"},
			},
			Required:    []string{"n", "from"},
			ExtraFields: map[string]any{"additionalProperties": false},
		},
	}
	guard := anthropic.BetaToolParam{
		Name:        "guard",
		Description: param.NewOpt("Write your guard: CEL returning an instruction or null. This ends your turn."),
		Strict:      param.NewOpt(true),
		InputSchema: anthropic.BetaToolInputSchemaParam{
			Properties:  map[string]any{"guard": map[string]any{"type": "string"}},
			Required:    []string{"guard"},
			ExtraFields: map[string]any{"additionalProperties": false},
		},
	}
	return []anthropic.BetaToolUnionParam{{OfTool: &expand}, {OfTool: &guard}}
}

// Guard asks Claude for a guard at the ask. Expansions are answered in the
// same conversation. If Claude ends without calling guard, its text is the
// guard, and the machine refuses it if it is not one.
func (m *ClaudeModel) Guard(a tree.AskModel) (string, error) {
	ctx := context.Background()
	msgs := []anthropic.BetaMessageParam{anthropic.NewBetaUserMessage(anthropic.NewBetaTextBlock(Render(a)))}
	expands := 0
	for {
		resp, err := m.Messages.New(ctx, anthropic.BetaMessageNewParams{
			Model:        m.Model,
			MaxTokens:    16000,
			System:       []anthropic.BetaTextBlockParam{{Text: claudeSystem, CacheControl: anthropic.NewBetaCacheControlEphemeralParam()}},
			Messages:     msgs,
			Tools:        m.tools(),
			OutputConfig: anthropic.BetaOutputConfigParam{Effort: m.Effort},
			Fallbacks:    anthropic.BetaFallbacksParamUnion{OfDefault: constant.ValueOf[constant.Default]()},
			Betas:        []anthropic.AnthropicBeta{anthropic.AnthropicBetaServerSideFallback2026_07_01},
		})
		if err != nil {
			return "", err
		}
		switch resp.StopReason {
		case anthropic.BetaStopReasonRefusal:
			return "", fmt.Errorf("the model declined: %s", resp.StopDetails.Explanation)
		case anthropic.BetaStopReasonMaxTokens:
			return "", errors.New("the model ran out of tokens before writing a guard")
		}
		msgs = append(msgs, resp.ToParam())
		var results []anthropic.BetaContentBlockParamUnion
		var text strings.Builder
		for _, block := range resp.Content {
			switch b := block.AsAny().(type) {
			case anthropic.BetaTextBlock:
				text.WriteString(b.Text)
			case anthropic.BetaToolUseBlock:
				switch b.Name {
				case "guard":
					var in struct{ Guard string }
					if err := json.Unmarshal([]byte(b.JSON.Input.Raw()), &in); err != nil {
						results = append(results, anthropic.NewBetaToolResultBlock(b.ID, "the input is not {\"guard\": string}", true))
						continue
					}
					return in.Guard, nil
				case "expand":
					var in struct{ N, From int }
					out, isErr := "", false
					if err := json.Unmarshal([]byte(b.JSON.Input.Raw()), &in); err != nil {
						out, isErr = "the input is not {\"n\": integer, \"from\": integer}", true
					} else if expands >= ExpandLimit {
						out, isErr = fmt.Sprintf("expanded %d times in this ask, the most allowed: write your guard", expands), true
					} else if out, err = Expand(a, in.N, in.From); err != nil {
						out, isErr = err.Error(), true
					} else {
						expands++
					}
					results = append(results, anthropic.NewBetaToolResultBlock(b.ID, out, isErr))
				default:
					results = append(results, anthropic.NewBetaToolResultBlock(b.ID, "no tool "+b.Name+": expand or guard", true))
				}
			}
		}
		if len(results) == 0 {
			return strings.TrimSpace(text.String()), nil
		}
		msgs = append(msgs, anthropic.NewBetaUserMessage(results...))
	}
}
