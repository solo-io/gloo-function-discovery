package source

// TODO(ashish)  - map to glue v1 api objects
type Upstream struct {
	ID        string
	Namespace string
	Name      string
	Type      string
	Functions []Function
	Spec      map[string]interface{}
}

type Function struct {
	Name string
	Spec map[string]interface{}
}

// FetcherFunc represents the function that knows how to discover
// functions for the given upstream
type FetcherFunc func(u *Upstream) ([]Function, error)

var (
	// FetcherRegistry maintains a list of discovery
	FetcherRegistry = make(map[string]FetcherFunc)
)
