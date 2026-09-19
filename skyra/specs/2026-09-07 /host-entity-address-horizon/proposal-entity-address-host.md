```go
type Entity struct {
    ID        EntityID
    Addresses []*EntityAddress
}

type Host struct {
    ID HostID
}

type HostAddress struct {
    Address string
    Host    *Host
}

type EntityAddress struct {
    Address       string
    HostAddresses []*HostAddress
}

type History struct{}

type LiveWork struct{}

func (a *EntityAddress) Representation(history *History, liveWork *LiveWork)
```
