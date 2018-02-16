package gcf

import (
	"context"
	"fmt"

	"golang.org/x/oauth2/google"

	"github.com/pkg/errors"
	"github.com/solo-io/glue-discovery/pkg/secret"
	"github.com/solo-io/glue-discovery/pkg/source"

	"google.golang.org/api/cloudfunctions/v1"
)

const (
	credentialKey   = "credential"
	projectIDKey    = "project"
	gcfUpstreamType = "gcf"
)

type gcfFetcher struct {
	secretRepo *secret.SecretRepo
}

func GetGCFFetcher(s *secret.SecretRepo) *gcfFetcher {
	return &gcfFetcher{secretRepo: s}
}

func (g *gcfFetcher) CanFetch(u *source.Upstream) bool {
	return u.Type == gcfUpstreamType
}

func (g *gcfFetcher) Fetch(u *source.Upstream) ([]source.Function, error) {
	// secretRef := secretRef(u)
	// data, exists := g.secretRepo.Get(secretRef)
	// if !exists {
	// 	return nil, fmt.Errorf("unable to get credential referenced by %s", secretRef)
	// }
	// fmt.Println(data)

	//ctx = oauth.NewServiceAccountFromFile(jsonfile)
	// jsonbytes, err := ioutil.ReadFile(serviceKeyFile)
	// if err != nil {
	// 	return nil, errors.Wrap(err, "unable to read servicekey")
	// }
	// ctx := oauth.NewServiceAccountFromKey(jsonbytes)
	ctx := context.Background()
	// relying on GOOGLE_APPLICATION_CREDENTIALS pointing to service key json file
	oauthClient, err := google.DefaultClient(ctx, cloudfunctions.CloudPlatformScope)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get OAuth client")
	}
	gcf, err := cloudfunctions.New(oauthClient)
	if err != nil {
		return nil, errors.Wrap(err, "unable to setup GCF service")
	}
	var functions []source.Function
	locationID := "-" // all locations
	parent := fmt.Sprintf("projects/%s/locations/%s", projectID(u), locationID)
	listCall := gcf.Projects.Locations.Functions.List(parent)
	//listCall.Fields("functions(name,status,versionId,entryPoint,httpsTrigger/*)")
	err = listCall.Pages(ctx, func(r *cloudfunctions.ListFunctionsResponse) error {
		// handle each page of list functions
		for _, f := range r.Functions {
			// limiting to active functions only; should we include functions
			// that are being deployed right now? not including; if deploy are
			// successful they should be active and added in the next poll period
			if f.Status == "ACTIVE" {
				var trigger map[string]string
				if f.HttpsTrigger != nil {
					trigger = map[string]string{
						"Type": "HTTP",
						"URL":  f.HttpsTrigger.Url,
					}
				} else {
					trigger = map[string]string{
						"Type":     "Event",
						"Event":    f.EventTrigger.EventType,
						"Resource": f.EventTrigger.Resource,
						"Service":  f.EventTrigger.Service,
					}
				}
				function := source.Function{
					Name: f.Name,
					Spec: map[string]interface{}{
						"Version": f.VersionId,
						"Entry":   f.EntryPoint,
						"Trigger": trigger,
					},
				}

				functions = append(functions, function)
			}
		}
		return nil
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to get list of GCF functions")
	}

	return functions, nil
}

func secretRef(u *source.Upstream) string {
	v, exists := u.Spec[credentialKey]
	if !exists {
		return ""
	}
	return u.Namespace + "/" + v.(string)
}

func projectID(u *source.Upstream) string {
	v, exists := u.Spec[projectIDKey]
	if !exists {
		return ""
	}
	return v.(string)
}
