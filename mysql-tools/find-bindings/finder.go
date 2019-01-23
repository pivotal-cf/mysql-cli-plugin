package find_bindings

import (
	"net/url"

	cfclient "github.com/cloudfoundry-community/go-cfclient"
	multierror "github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
)

//go:generate counterfeiter . CFClient
type CFClient interface {
	GetAppByGuid(guid string) (cfclient.App, error)
	GetOrgByGuid(spaceGUID string) (cfclient.Org, error)
	GetSpaceByGuid(spaceGUID string) (cfclient.Space, error)
	ListServicesByQuery(url.Values) ([]cfclient.Service, error)
	ListServiceBindingsByQuery(query url.Values) ([]cfclient.ServiceBinding, error)
	ListServicePlansByQuery(query url.Values) ([]cfclient.ServicePlan, error)
	ListServiceKeysByQuery(query url.Values) ([]cfclient.ServiceKey, error)
	ListServiceInstancesByQuery(query url.Values) ([]cfclient.ServiceInstance, error)
}

type Binding struct {
	Name                string
	ServiceInstanceName string
	ServiceInstanceGuid string
	OrgName             string
	SpaceName           string
	Type                string
}

type bindingFinder struct {
	cfClient CFClient
}

func NewBindingFinder(cfClient CFClient) *bindingFinder {
	return &bindingFinder{
		cfClient: cfClient,
	}
}

func (bf *bindingFinder) FindBindings(serviceLabel string) ([]Binding, error) {
	//cf curl " /v2/services?q=label:p.mysql"
	//	get resources entity.service_plans_url "/v2/services/9cbbd018-236f-4171-8585-594ebfde52f2/service_plans"
	//	cf curl service_plans_url
	//		get resources entity.service_instances_url "/v2/spaces/8b892a65-bf0e-4276-ad47-30757c4f2251/service_instances"
	//		cf curl service_instances_url
	//			get resources entity.service_bindings_url "/v2/service_instances/00d4ce31-bbbe-48f6-b15d-fcbd3380f50a/service_bindings"
	//			get resources entity.service_keys_url "/v2/service_instances/00d4ce31-bbbe-48f6-b15d-fcbd3380f50a/service_keys"
	//			cf curl service_bindings_url
	//				get resources entity.app_guid
	//			cf curl service_keys_url
	//				get resources entity.name    ?
	//				get resources metadata.guid  ?
	//			get resources entity.space_url "/v2/spaces/8b892a65-bf0e-4276-ad47-30757c4f2251"
	//			cf curl space_url
	//				get resources entity.name
	//				get resources entity.organization_url "/v2/organizations/10b9207b-1c15-46d8-9946-a2374b8c40e5"
	//				cf curl organization_url
	//					get resources entity.name
	//u := url.Values{}
	//u.Set("q", fmt.Sprintf("label:%s", serviceLabel))
	//services, _ := bf.cfClient.ListServicesByQuery(u)
	//return []ServiceBinding{
	//	{App: "app", Key: "", Org: "org-name", Space: "space-name"},
	//}, nil
	serviceGUID, err := bf.serviceGUIDForLabel(serviceLabel)
	if err != nil {
		return nil, errors.Wrapf(err, `failed to lookup service matching label %q`, serviceLabel)
	}

	servicePlans, err := bf.servicePlansForServiceGUID(serviceGUID)
	if err != nil {
		return nil, errors.Wrapf(err, `failed to lookup service plans for service (guid: %q, label: %q)`, serviceGUID, serviceLabel)
	}

	var (
		result []Binding
		errs   error
	)

	serviceInstances, err := bf.serviceInstancesForServicePlans(servicePlans)
	if err != nil {
		errs = multierror.Append(errs, err)
	}

	for _, instance := range serviceInstances {
		bindings, err := bf.listServiceBindingsForInstance(instance)
		if err != nil {
			errs = multierror.Append(errs, err)
		}
		result = append(result, bindings...)

		bindings, err = bf.listServiceKeysForInstance(instance)
		if err != nil {
			errs = multierror.Append(errs, err)
		}
		result = append(result, bindings...)
	}

	return result, errs
}

func (bf *bindingFinder) serviceGUIDForLabel(serviceLabel string) (serviceGUID string, err error) {
	query := url.Values{}
	query.Set("q", "label:"+serviceLabel)
	services, err := bf.cfClient.ListServicesByQuery(query)
	if err != nil {
		return "", err
	}

	switch len(services) {
	case 0:
		return "", errors.New("no matching services found")
	case 1:
		return services[0].Guid, nil
	default:
		return "", errors.Errorf("found %d matching services, expected 1", len(services))
	}
}

func (bf *bindingFinder) servicePlansForServiceGUID(serviceGUID string) (servicePlans []cfclient.ServicePlan, err error) {
	query := url.Values{}
	query.Set("q", "service_guid:"+serviceGUID)
	return bf.cfClient.ListServicePlansByQuery(query)
}

func (bf *bindingFinder) serviceInstancesForServicePlans(servicePlans []cfclient.ServicePlan) ([]cfclient.ServiceInstance, error) {
	var (
		result []cfclient.ServiceInstance
		errs   error
	)

	for _, plan := range servicePlans {
		instances, err := bf.cfClient.ListServiceInstancesByQuery(url.Values{
			"q": []string{
				"service_plan_guid:" + plan.Guid,
			},
		})
		if err != nil {
			errs = multierror.Append(
				errs,
				errors.Wrapf(err, `failed to lookup service instances for service plan (name: %q, guid: %q)`, plan.Name, plan.Guid),
			)
		}

		result = append(result, instances...)
	}

	return result, errs
}

func (bf *bindingFinder) listServiceBindingsForInstance(instance cfclient.ServiceInstance) ([]Binding, error) {
	query := url.Values{}
	query.Set("q", "service_instance_guid:"+instance.Guid)
	serviceBindings, err := bf.cfClient.ListServiceBindingsByQuery(query)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to retrieve service bindings for service instance (name: %q guid: %q)", instance.Name, instance.Guid)
	}

	var (
		result []Binding
		errs   error
	)

	for _, b := range serviceBindings {
		if err != nil {
			errs = multierror.Append(errs, errors.Wrapf(err, "failed to resolve secure binding credentials for app (guid: %q)", b.AppGuid))
			continue
		}

		app, err := bf.cfClient.GetAppByGuid(b.AppGuid)
		if err != nil {
			errs = multierror.Append(errs, errors.Wrapf(err, "failed to lookup app info for app guid %q", b.AppGuid))
			continue
		}

		result = append(result, Binding{
			Name:                app.Name,
			ServiceInstanceName: instance.Name,
			ServiceInstanceGuid: instance.Guid,
			OrgName:             app.SpaceData.Entity.OrgData.Entity.Name,
			SpaceName:           app.SpaceData.Entity.Name,
			Type:                "AppBinding",
		})
	}

	return result, errs
}

func (bf *bindingFinder) listServiceKeysForInstance(instance cfclient.ServiceInstance) ([]Binding, error) {
	query := url.Values{}
	query.Set("q", "service_instance_guid:"+instance.Guid)
	serviceKeys, err := bf.cfClient.ListServiceKeysByQuery(query)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to retrieve service keys for service instance (name: %q guid: %q)", instance.Name, instance.Guid)
	}

	var (
		result []Binding
		errs   error
	)

	for _, k := range serviceKeys {
		space, err := bf.spaceDataForGUID(instance.SpaceGuid)
		if err != nil {
			// XXX: Think about the error and what information to surface here
			errs = multierror.Append(errs, errors.Wrapf(err, "failed to lookup space info for service key (name: %q instance-guid: %q)", k.Name, instance.Guid))
			continue
		}

		result = append(result, Binding{
			Name:                k.Name,
			ServiceInstanceName: instance.Name,
			ServiceInstanceGuid: instance.Guid,
			OrgName:             space.OrgData.Entity.Name,
			SpaceName:           space.Name,
			Type:                "ServiceKeyBinding",
		})
	}

	return result, errs
}

func (bf *bindingFinder) spaceDataForGUID(spaceGUID string) (cfclient.Space, error) {
	space, err := bf.cfClient.GetSpaceByGuid(spaceGUID)
	if err != nil {
		return space, err
	}

	org, err := bf.cfClient.GetOrgByGuid(space.OrganizationGuid)
	if err != nil {
		return space, err
	}

	space.OrgData.Entity = org

	return space, nil
}
