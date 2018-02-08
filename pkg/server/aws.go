package server

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/pkg/errors"
	"github.com/solo-io/glue-discovery/pkg/source/aws"
	apiv1 "github.com/solo-io/glue/pkg/api/types/v1"
	solov1 "github.com/solo-io/glue/pkg/platform/kube/crd/solo.io/v1"
)

// adapter between aws poller and what controller expects
// can be removed if we make aws poller implement necessary
// methods; not doing for now since aws poller doesn't
// need to be aware of any data type outside the package
type awsHandler struct {
	controller *controller
	poller     *aws.AWSPoller
}

func newAWSHandler(c *controller) awsHandler {
	updater := func(r aws.Region) error {
		upstream, exists, err := c.get(r.ID)
		if err != nil {
			return errors.Wrapf(err, "Unable to update upstream %s", r.ID)
		}
		if !exists {
			log.Printf("upstream %s not found, will not update", r.ID)
			return nil
		}
		upstream.Spec.Functions = toFunctions(r.Lambdas)
		log.Println("updating upstream ", r.ID)
		return c.set(upstream)
	}
	poller := aws.NewAWSPoller(aws.AWSFetcher, updater)
	return awsHandler{controller: c, poller: poller}
}

func (a awsHandler) Update(u *solov1.Upstream) {
	a.poller.AddUpdateRegion(toRegion(u))
}

func (a awsHandler) Remove(u *solov1.Upstream) {
	a.poller.RemoveRegion(toID(u))
}

func (a awsHandler) Start(stop chan struct{}) {
	a.poller.Start(1*time.Minute, stop)
}

func toRegion(u *solov1.Upstream) aws.Region {
	r := aws.Region{
		ID:      toID(u),
		Name:    u.Spec.Spec["region"].(string),
		Token:   toToken(""), //u.Spec.Spec["token"].(string)),
		Lambdas: toLambdas(u.Spec.Functions),
	}
	return r
}

func toID(u *solov1.Upstream) string {
	return fmt.Sprintf("%s/%s", u.Namespace, u.Name)
}
func toToken(s string) aws.AccessToken {
	return aws.AccessToken{
		ID:     os.Getenv("AWS_ACCESS_KEY_ID"),
		Secret: os.Getenv("AWS_SECRET_ACCESS_KEY"),
	}
}

func toLambdas(functions []apiv1.Function) []aws.Lambda {
	lambdas := make([]aws.Lambda, len(functions))
	for i, f := range functions {
		lambdas[i] = aws.Lambda{
			Name:      f.Spec["FunctionName"].(string),
			Qualifier: f.Spec["Qualifier"].(string),
		}
	}
	return lambdas
}

func toFunctions(lambdas []aws.Lambda) []apiv1.Function {
	functions := make([]apiv1.Function, len(lambdas))
	for i, l := range lambdas {
		functions[i] = apiv1.Function{
			Name: l.Name + ":" + l.Qualifier,
			Spec: map[string]interface{}{
				"FunctionName": l.Name,
				"Qualifier":    l.Qualifier,
			},
		}
	}
	return functions
}
