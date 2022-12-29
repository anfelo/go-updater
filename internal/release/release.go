package release

// Release - the release model
type Release struct {
	ID         string
	URL        string
	Draft      bool
	Prerelease bool
	Body       string
	Name       string
	TagName    string
}
