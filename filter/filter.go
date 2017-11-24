package filter

type Request interface {
    HostName() string
    Path() string
    Port() int
}
