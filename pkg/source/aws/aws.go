package aws

import (
	"log"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"
)

// TODO(ashish) - when adding next source abstract this
// to generic poller that takes fetchers and updaters

// AccessToken used to talk to AWS
type AccessToken struct {
	ID     string
	Secret string
}

// Region represents the AWS region and Lambdas in the region
type Region struct {
	ID      string
	Name    string
	Token   AccessToken
	Lambdas []Lambda
}

// Lambda represents AWS Lambda, each qualifier is treated as separate Lambda
type Lambda struct {
	Name      string
	Qualifier string
}

// Fetcher gets collection of Lambdas for given region
type Fetcher func(string, AccessToken) ([]Lambda, error)

// Updater updates changes in Lambdas; for example saves
// it in CRDs
type Updater func(Region) error

type awsPoller struct {
	repo    *memRepo
	fetcher Fetcher
	updater Updater
}

func NewAWSPoller(f Fetcher, u Updater) *awsPoller {
	return &awsPoller{
		repo:    newRepo(),
		fetcher: f,
		updater: u,
	}
}

func (a *awsPoller) AddUpdateRegion(region Region) {
	existing, exists := a.repo.get(region.ID)
	if exists {
		region.Lambdas = existing.Lambdas
	}
	a.repo.set(region)
}

func (a *awsPoller) RemoveRegion(regionName string) {
	a.repo.delete(regionName)
}

func (a *awsPoller) Start(pollPeriod time.Duration, stop chan struct{}) {
	go func() { wait.Until(a.run, pollPeriod, stop) }()
}

func (a *awsPoller) run() {
	regions := a.repo.regions()
	for _, r := range regions {
		newLambdas, err := a.fetcher(r.Name, r.Token)
		if err != nil {
			log.Printf("Unable to get lambdas for %s, %q\n", r.Name, err)
			continue
		}

		if diff(newLambdas, r.Lambdas) {
			updated := Region{
				ID:      r.ID,
				Name:    r.Name,
				Token:   r.Token,
				Lambdas: newLambdas}
			if err := a.updater(updated); err != nil {
				log.Printf("unable to update change in Lambdas for %s: %q\n", r.Name, err)
				continue
			}
			a.repo.set(updated)
		}
	}
}

func diff(l, r []Lambda) bool {
	if len(l) != len(r) {
		return true
	}

	m := make(map[string]bool, len(l))
	for _, li := range l {
		m[li.Name+":"+li.Qualifier] = true
	}

	for _, ri := range r {
		_, exists := m[ri.Name+":"+ri.Qualifier]
		if !exists {
			return true
		}
	}

	return false
}
