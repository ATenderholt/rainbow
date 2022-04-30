package domain

type MotoRequest struct {
	ID            int64
	Service       string
	Method        string
	Path          string
	Authorization string
	ContentType   string
	Target        string
	Payload       string
}
