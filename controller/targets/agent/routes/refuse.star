# Refusing is a route: the agent sees echo's return as a fault.
def route():
    return skyra.fault("not today")
