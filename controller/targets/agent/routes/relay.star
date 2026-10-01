# The first time echo answers, send its answer back to echo, louder; the
# second time (there is a previous return of this type), wake the agent.
def route():
    if skyra.previous() == None:
        return skyra.send("echo", {"text": skyra.get("$") + "!"})
    return skyra.interrupt()
