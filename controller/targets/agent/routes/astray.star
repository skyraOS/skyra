# A send to a name outside the agent's reach is refused.
def route():
    return skyra.send("shell", {"command": "ls"})
