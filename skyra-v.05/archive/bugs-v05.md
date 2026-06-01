# bugs

## ~ref enforcement propagation

### issue

when a being crosses exchanges without ~ref, exchange detects the violation and returns an error string. but that string flows forward through the system as a normal response — thread routes it, exchange records it as a conversation entry, and the target being receives the raw error as a message.

this happens because exchange runs downstream of act. by the time exchange fires, act has already returned. there's no path for the error to reach act's retry loop.

### solution

exchange sets an error reality on the relation instead of returning an error string. it returns empty.

```
r.Realities["error"] = &Error{Message: "..."}
return ""
```

self checks for the error after the full stack returns. if there's an error, self feeds it into act's present as a warning and re-fires act. act sees the warning and retries with ~ref.

each layer owns its own enforcement and its own errors. exchange doesn't need to know about act's retry loop. self routes the error. act corrects.

this pattern generalizes — any reality can surface an error via `r.Realities["error"]`, and the layer above decides how to handle it.
