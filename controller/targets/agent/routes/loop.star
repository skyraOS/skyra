# Always send on: the chain ends at its depth, and the agent is woken.
def route():
    return skyra.send("echo", {"text": skyra.get("$") + "!"})
