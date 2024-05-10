package templates

type Heading struct {
	Key   string
	Count string
}

type Links struct {
	InternalTotal int
	ExternalTotal int

	InaccessibleTotal int
}

type Details struct {
	URL          string
	Title        string
	HTMLVersion  string
	Headings     []Heading
	Links        Links
	HasLoginForm bool
}
