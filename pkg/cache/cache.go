package cache

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"provisioner/api/v1alpha1"
)

const (
	ByMAC = "spec.macAddress"
	ByIP  = "spec.ipAddress"
)

type Cache struct {
	cache cache.Cache
}

func New(ctx context.Context, config *rest.Config) (_ *Cache, err error) {
	c := &Cache{}

	c.cache, err = cache.New(config, cache.Options{})
	if err != nil {
		return nil, err
	}

	if _, err = c.cache.GetInformer(ctx, &v1alpha1.Machine{}); err != nil {
		return nil, err
	}

	if err = c.cache.IndexField(ctx, &v1alpha1.Machine{}, ByMAC, func(o client.Object) []string {
		return []string{o.(*v1alpha1.Machine).Spec.MacAddress}
	}); err != nil {
		return nil, err
	}

	if err = c.cache.IndexField(ctx, &v1alpha1.Machine{}, ByIP, func(o client.Object) []string {
		return []string{o.(*v1alpha1.Machine).Spec.IPAddress}
	}); err != nil {
		return nil, err
	}

	go func() {
		panic(c.cache.Start(ctx))
	}()

	if !c.cache.WaitForCacheSync(ctx) {
		return nil, fmt.Errorf("could not sync")
	}

	return c, nil
}

func (c *Cache) Get(ctx context.Context, field, value string) (*v1alpha1.Machine, error) {
	set := fields.Set{field: value}

	l := &v1alpha1.MachineList{}
	err := c.cache.List(ctx, l, client.MatchingFieldsSelector{Selector: set.AsSelector()}, client.Limit(2))
	if err != nil {
		return nil, err
	}

	if len(l.Items) != 1 {
		return nil, fmt.Errorf("%d items found for %q", len(l.Items), set)
	}

	return &l.Items[0], nil
}
