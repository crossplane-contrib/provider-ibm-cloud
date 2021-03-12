package resourcekey

import (
	"testing"

	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/reference"
	"github.com/go-openapi/strfmt"
	"github.com/google/go-cmp/cmp"

	rcv2 "github.com/IBM/platform-services-go-sdk/resourcecontrollerv2"

	"github.com/crossplane-contrib/provider-ibm-cloud/apis/resourcecontrollerv2/v1alpha1"
	ibmc "github.com/crossplane-contrib/provider-ibm-cloud/pkg/clients"
)

var (
	role                                           = "Manager"
	role2                                          = "Reader"
	id                                             = "crn:v1:bluemix:public:cloud-object-storage:global:a/0b5a00334eaf9eb9339d2ab48f20d7f5:f931e669-6c11-4d4d-b720-8b2f844a6d9e:resource-key:bbeca5fe-283f-443c-9aca-cd3f72c6f493"
	createdBy                                      = "user00001"
	iamCompatible                                  = true
	keyName                                        = "cos-key"
	resourceGroupID                                = "mock-resource-group-id"
	resInstURL                                     = "/v2/resource_keys/614566d9-7ae6-4755-a5ae-83a8dd806ee4"
	sourceCRN                                      = "crn:v1:bluemix:public:cloud-object-storage:global:a/0b5a00334eaf9eb9339d2ab48f20d7f5:78d88b2b-bbbb-aaaa-8888-5c26e8b6a555::"
	resCRN                                         = "crn:v1:bluemix:public:key:global:a/0b5a00334eaf9eb9339d2ab48f20d7f5:78d88b2b-bbbb-aaaa-8888-5c26e8b6a555::"
	accountID                                      = "fake-account-id"
	guid                                           = "78d88b2b-bbbb-aaaa-8888-5c26e8b6a555"
	createdAt, _                                   = strfmt.ParseDateTime("2020-10-31T02:33:06Z")
	state                                          = StateActive
	url                                            = "/v2/resource_keys/614566d9-7ae6-4755-a5ae-83a8dd806ee4"
	apikey                                         = "fake_api_key"
	iamApikeyDescr                                 = "Auto-generated for key 055aa817-cb92-4a2a-8155-0f7588b2e844"
	iamApikeyName                                  = "cos-key"
	iamRoleCRN                                     = "crn:v1:bluemix:public:iam::::serviceRole:Manager"
	iamServiceidCRN                                = "crn:v1:bluemix:public:iam-identity::a/0b5a00334eaf9eb9339d2ab48f20d7f5::serviceid:ServiceId-ed663a94-5a2e-487b-a26d-2c7811565790"
	connectionCliArguments00                       = "host=fake-host.databases.appdomain.cloud port=31700 dbname=ibmclouddb user=fake-user sslmode=verify-full"
	connectionCliBin                               = "psql"
	connectionCliCertificateCertificateBase64      = "ZmFrZS1jZXJ0aWZpY2F0ZQo="
	connectionCliCertificateName                   = "fake-cert-id"
	connectionCliComposed0                         = "PGPASSWORD=fake-password PGSSLROOTCERT=fake-cert-id psql 'host=fake-host.databases.appdomain.cloud port=31700 dbname=ibmclouddb user=fake-user sslmode=verify-full'"
	connectionCliEnvironmentPgpassword             = "fake-password"
	connectionCliEnvironmentPgsslrootcert          = "fake-cert-id"
	connectionCliType                              = "cli"
	connectionPostgresAuthenticationMethod         = "direct"
	connectionPostgresAuthenticationPassword       = "fake-password"
	connectionPostgresAuthenticationUsername       = "fake-user"
	connectionPostgresCertificateCertificateBase64 = "ZmFrZS1jZXJ0aWZpY2F0ZQo="
	connectionPostgresCertificateName              = "fake-cert-id"
	connectionPostgresComposed0                    = "postgres://fake-user:fake-password@fake-host.databases.appdomain.cloud:31700/ibmclouddb?sslmode=verify-full"
	connectionPostgresDatabase                     = "ibmclouddb"
	connectionPostgresHosts0Hostname               = "fake-host.databases.appdomain.cloud"
	connectionPostgresHosts0Port                   = "31700"
	connectionPostgresPath                         = "/ibmclouddb"
	connectionPostgresQueryOptionsSslmode          = "verify-full"
	connectionPostgresScheme                       = "postgres"
	connectionPostgresType                         = "uri"
	instanceAdministrationAPIDeploymentID          = "crn:v1:bluemix:public:databases-for-postgresql:us-south:a/fake-id::"
	instanceAdministrationAPIInstanceID            = "crn:v1:bluemix:public:databases-for-postgresql:us-south:a/fake-id::"
	instanceAdministrationAPIRoot                  = "https://api.us-south.databases.cloud.ibm.com/v5/ibm"
	endpoints                                      = "https://control.cloud-object-storage.cloud.ibm.com/v2/endpoints"
	resourceInstanceID                             = "crn:v1:bluemix:public:cloud-object-storage:global:a/0b5a00334eaf9eb9339d2ab48f20d7f5:f931e669-6c11-4d4d-b720-8b2f844a6d9e::"
	creds1AddProps                                 = map[string]interface{}{
		"endpoints":            endpoints,
		"resource_instance_id": resourceInstanceID,
	}
	creds2AddProps = map[string]interface{}{
		"connection": map[string]interface{}{
			"cli": map[string]interface{}{
				"arguments": []interface{}{
					[]interface{}{connectionCliArguments00},
				},
				"bin": connectionCliBin,
				"certificate": map[string]interface{}{
					"certificate_base64": connectionCliCertificateCertificateBase64,
					"name":               connectionCliCertificateName,
				},
				"composed": []interface{}{connectionCliComposed0},
				"environment": map[string]interface{}{
					"PGPASSWORD":    connectionCliEnvironmentPgpassword,
					"PGSSLROOTCERT": connectionCliEnvironmentPgsslrootcert,
				},
				"type": connectionCliType,
			},
			"postgres": map[string]interface{}{
				"authentication": map[string]interface{}{
					"method":   connectionPostgresAuthenticationMethod,
					"password": connectionPostgresAuthenticationPassword,
					"username": connectionPostgresAuthenticationUsername,
				},
				"certificate": map[string]interface{}{
					"certificate_base64": connectionCliCertificateCertificateBase64,
					"name":               connectionCliCertificateName,
				},
				"composed": []interface{}{connectionPostgresComposed0},
				"database": connectionPostgresDatabase,
				"hosts": []interface{}{map[string]interface{}{
					"hostname": connectionPostgresHosts0Hostname,
					"port":     connectionPostgresHosts0Port,
				}},
				"path": connectionPostgresPath,
				"query_options": map[string]interface{}{
					"sslmode": connectionPostgresQueryOptionsSslmode,
				},
				"scheme": connectionPostgresScheme,
				"type":   connectionPostgresType,
			},
		},
		"instance_administration_api": map[string]interface{}{
			"deployment_id": instanceAdministrationAPIDeploymentID,
			"instance_id":   instanceAdministrationAPIInstanceID,
			"root":          instanceAdministrationAPIRoot,
		},
	}
)

func params(m ...func(*v1alpha1.ResourceKeyParameters)) *v1alpha1.ResourceKeyParameters {
	p := &v1alpha1.ResourceKeyParameters{
		Name:   keyName,
		Role:   &role,
		Source: &sourceCRN,
	}

	for _, f := range m {
		f(p)
	}
	return p
}

func observation(m ...func(*v1alpha1.ResourceKeyObservation)) *v1alpha1.ResourceKeyObservation {
	o := &v1alpha1.ResourceKeyObservation{
		AccountID:           accountID,
		CreatedBy:           createdBy,
		DeletedBy:           "",
		IamCompatible:       iamCompatible,
		CreatedAt:           GenerateMetaV1Time(&createdAt),
		CRN:                 resCRN,
		DeletedAt:           nil,
		GUID:                guid,
		ID:                  id,
		ResourceGroupID:     resourceGroupID,
		ResourceInstanceURL: resInstURL,
		State:               state,
		URL:                 url,
		UpdatedAt:           GenerateMetaV1Time(&createdAt),
		UpdatedBy:           createdBy,
	}

	for _, f := range m {
		f(o)
	}
	return o
}

func instance(m ...func(*rcv2.ResourceKey)) *rcv2.ResourceKey {
	i := &rcv2.ResourceKey{
		AccountID: &accountID,
		CreatedAt: &createdAt,
		CreatedBy: &createdBy,
		CRN:       &resCRN,
		DeletedAt: nil,
		DeletedBy: nil,
		GUID:      &guid,
		ID:        &id,
		Name:      &keyName,
		// Parameters:    parameters, // TODO
		ResourceGroupID:     &resourceGroupID,
		State:               &state,
		URL:                 &url,
		UpdatedAt:           &createdAt,
		UpdatedBy:           &createdBy,
		IamCompatible:       &iamCompatible,
		ResourceInstanceURL: &resInstURL,
		Role:                &role,
		SourceCRN:           &sourceCRN,
	}

	for _, f := range m {
		f(i)
	}
	return i
}

func instanceOpts(m ...func(*rcv2.CreateResourceKeyOptions)) *rcv2.CreateResourceKeyOptions {
	i := &rcv2.CreateResourceKeyOptions{
		Name: &keyName,
		//Parameters:     parameters,
		Role:   &role,
		Source: &sourceCRN,
	}
	for _, f := range m {
		f(i)
	}
	return i
}

func instanceUpdOpts(m ...func(*rcv2.UpdateResourceKeyOptions)) *rcv2.UpdateResourceKeyOptions {
	i := &rcv2.UpdateResourceKeyOptions{
		ID:   &id,
		Name: &keyName,
	}

	for _, f := range m {
		f(i)
	}
	return i
}

func credentials(m ...func(*rcv2.Credentials)) *rcv2.Credentials {
	i := &rcv2.Credentials{
		Apikey:               &apikey,
		IamApikeyDescription: &iamApikeyDescr,
		IamApikeyName:        &iamApikeyName,
		IamRoleCRN:           &iamRoleCRN,
		IamServiceidCRN:      &iamServiceidCRN,
	}
	for _, f := range m {
		f(i)
	}
	return i
}

func cr(m ...func(*v1alpha1.ResourceKey)) *v1alpha1.ResourceKey {
	i := &v1alpha1.ResourceKey{}
	for _, f := range m {
		f(i)
	}
	return i
}

func TestGenerateCreateResourceKeyOptions(t *testing.T) {
	type args struct {
		name   string
		params v1alpha1.ResourceKeyParameters
	}
	type want struct {
		instance *rcv2.CreateResourceKeyOptions
	}
	cases := map[string]struct {
		args args
		want want
	}{
		"FullConversion": {
			args: args{name: keyName, params: *params()},
			want: want{instance: instanceOpts()},
		},
		"MissingFields": {
			args: args{
				name: keyName,
				params: *params(func(p *v1alpha1.ResourceKeyParameters) {
					p.Role = nil
				})},
			want: want{instance: instanceOpts(func(rk *rcv2.CreateResourceKeyOptions) {
				rk.Role = nil
			})},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {

			r := &rcv2.CreateResourceKeyOptions{}
			GenerateCreateResourceKeyOptions(tc.args.name, tc.args.params, r)
			if diff := cmp.Diff(tc.want.instance, r); diff != "" {
				t.Errorf("GenerateCreateResourceKeyOptions(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestGenerateUpdateResourceKeyOptions(t *testing.T) {
	type args struct {
		name   string
		params v1alpha1.ResourceKeyParameters
	}
	type want struct {
		instance *rcv2.UpdateResourceKeyOptions
	}
	cases := map[string]struct {
		args args
		want want
	}{
		"FullConversion": {
			args: args{name: keyName, params: *params()},
			want: want{instance: instanceUpdOpts()},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			r := &rcv2.UpdateResourceKeyOptions{}
			GenerateUpdateResourceKeyOptions(tc.args.name, id, tc.args.params, r)
			if diff := cmp.Diff(tc.want.instance, r); diff != "" {
				t.Errorf("GenerateUpdateResourceKeyOptions(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestLateInitializeSpecs(t *testing.T) {
	type args struct {
		instance *rcv2.ResourceKey
		params   *v1alpha1.ResourceKeyParameters
	}
	type want struct {
		params *v1alpha1.ResourceKeyParameters
	}
	cases := map[string]struct {
		args args
		want want
	}{
		"SomeFields": {
			args: args{
				params: params(func(p *v1alpha1.ResourceKeyParameters) {
					p.Role = nil
				}),
				instance: instance(func(i *rcv2.ResourceKey) {
					i.Role = &role
				}),
			},
			want: want{
				params: params(func(p *v1alpha1.ResourceKeyParameters) {
					p.Role = &role
					p.Source = &sourceCRN
				})},
		},
		"AllFilledAlready": {
			args: args{
				params:   params(),
				instance: instance(),
			},
			want: want{
				params: params()},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			LateInitializeSpec(tc.args.params, tc.args.instance)
			if diff := cmp.Diff(tc.want.params, tc.args.params); diff != "" {
				t.Errorf("LateInitializeSpec(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestGenerateObservation(t *testing.T) {
	type args struct {
		instance *rcv2.ResourceKey
	}
	type want struct {
		obs v1alpha1.ResourceKeyObservation
	}
	cases := map[string]struct {
		args args
		want want
	}{
		"FullConversion": {
			args: args{
				instance: instance(func(p *rcv2.ResourceKey) {
				}),
			},
			want: want{*observation()},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			o, err := GenerateObservation(tc.args.instance)
			if diff := cmp.Diff(nil, err); diff != "" {
				t.Errorf("GenerateObservation(...): want error != got error:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.obs, o); diff != "" {
				t.Errorf("GenerateObservation(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestIsUpToDate(t *testing.T) {
	type args struct {
		params   *v1alpha1.ResourceKeyParameters
		instance *rcv2.ResourceKey
	}
	type want struct {
		upToDate bool
		isErr    bool
	}
	cases := map[string]struct {
		args args
		want want
	}{
		"IsUpToDate": {
			args: args{
				params:   params(),
				instance: instance(),
			},
			want: want{upToDate: true, isErr: false},
		},
		"NeedsUpdate": {
			args: args{
				params: params(),
				instance: instance(func(i *rcv2.ResourceKey) {
					i.Role = &role2
				}),
			},
			want: want{upToDate: false, isErr: false},
		},
		"NeedsUpdateOnName": {
			args: args{
				params: params(),
				instance: instance(func(i *rcv2.ResourceKey) {
					i.Name = reference.ToPtrValue("different-name")
				}),
			},
			want: want{upToDate: false, isErr: false},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {

			r, err := IsUpToDate(keyName, tc.args.params, tc.args.instance, logging.NewNopLogger())
			if err != nil && !tc.want.isErr {
				t.Error("IsUpToDate(...) unexpected error")
			}
			if diff := cmp.Diff(tc.want.upToDate, r); diff != "" {
				t.Errorf("IsUpToDate(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestGetConnectionDetails(t *testing.T) {
	type args struct {
		cr       *v1alpha1.ResourceKey
		instance *rcv2.ResourceKey
	}
	type want struct {
		conn managed.ConnectionDetails
	}
	cases := map[string]struct {
		args args
		want want
	}{
		"SimpleCredsWithTemplate": {
			args: args{
				cr: cr(func(c *v1alpha1.ResourceKey) {
					c.Spec.ConnectionTemplates = map[string]string{
						"apikey":    `{{ .apikey }}`,
						"endpoints": `{{ .endpoints }}`,
					}
				}),
				instance: instance(func(p *rcv2.ResourceKey) {
					p.Credentials = credentials(func(r *rcv2.Credentials) {
						for k, v := range creds1AddProps {
							r.SetProperty(k, v)
						}
					})
				}),
			},
			want: want{managed.ConnectionDetails{
				"apikey":    []byte(apikey),
				"endpoints": ibmc.ConvertVarsMap(creds1AddProps)["endpoints"],
			}},
		},
		"ComplexCredsWithTemplate": {
			args: args{
				cr: cr(func(c *v1alpha1.ResourceKey) {
					c.Spec.ConnectionTemplates = map[string]string{
						"connectionString": `jdbc:postgresql://{{ (index .connection.postgres.hosts 0).hostname }}:{{ (index .connection.postgres.hosts 0).port }}{{ .connection.postgres.path }}`,
						"user":             `{{ .connection.postgres.authentication.username }}`,
						"password":         `{{ .connection.postgres.authentication.password }}`,
					}
				}),
				instance: instance(func(p *rcv2.ResourceKey) {
					p.Credentials = credentials(func(r *rcv2.Credentials) {
						r.Apikey = nil
						r.IamApikeyDescription = nil
						r.IamApikeyName = nil
						r.IamRoleCRN = nil
						r.IamServiceidCRN = nil
						for k, v := range creds2AddProps {
							r.SetProperty(k, v)
						}
					})
				}),
			},
			want: want{managed.ConnectionDetails{
				"password":         []byte(connectionPostgresAuthenticationPassword),
				"user":             []byte(connectionPostgresAuthenticationUsername),
				"connectionString": []byte("jdbc:postgresql://" + connectionPostgresHosts0Hostname + ":" + connectionPostgresHosts0Port + connectionPostgresPath),
			}},
		},
		"SimpleCredsNoTemplate": {
			args: args{
				cr: cr(),
				instance: instance(func(p *rcv2.ResourceKey) {
					p.Credentials = credentials(func(r *rcv2.Credentials) {
						for k, v := range creds1AddProps {
							r.SetProperty(k, v)
						}
					})
				}),
			},
			want: want{managed.ConnectionDetails{
				"apikey":               []byte(apikey),
				"iamApikeyName":        []byte(iamApikeyName),
				"endpoints":            ibmc.ConvertVarsMap(creds1AddProps)["endpoints"],
				"iamApikeyDescription": []byte(iamApikeyDescr),
				"iamRoleCRN":           []byte(iamRoleCRN),
				"iamServiceidCRN":      []byte(iamServiceidCRN),
				"resource_instance_id": ibmc.ConvertVarsMap(creds1AddProps)["resource_instance_id"],
			}},
		},
		"ComplexCredsNoTemplate": {
			args: args{
				cr: cr(),
				instance: instance(func(p *rcv2.ResourceKey) {
					p.Credentials = credentials(func(r *rcv2.Credentials) {
						r.Apikey = nil
						r.IamApikeyDescription = nil
						r.IamApikeyName = nil
						r.IamRoleCRN = nil
						r.IamServiceidCRN = nil
						for k, v := range creds2AddProps {
							r.SetProperty(k, v)
						}
					})
				}),
			},
			want: want{managed.ConnectionDetails{
				"apikey":                       nil,
				"iamApikeyDescription":         nil,
				"iamApikeyName":                nil,
				"iamRoleCRN":                   nil,
				"iamServiceidCRN":              nil,
				"connection.cli.arguments.0.0": []byte(connectionCliArguments00),
				"connection.cli.bin":           []byte(connectionCliBin),
				"connection.cli.certificate.certificate_base64":      []byte(connectionCliCertificateCertificateBase64),
				"connection.cli.certificate.name":                    []byte(connectionCliCertificateName),
				"connection.cli.composed.0":                          []byte(connectionCliComposed0),
				"connection.cli.environment.PGPASSWORD":              []byte(connectionCliEnvironmentPgpassword),
				"connection.cli.environment.PGSSLROOTCERT":           []byte(connectionCliEnvironmentPgsslrootcert),
				"connection.cli.type":                                []byte(connectionCliType),
				"connection.postgres.authentication.method":          []byte(connectionPostgresAuthenticationMethod),
				"connection.postgres.authentication.password":        []byte(connectionPostgresAuthenticationPassword),
				"connection.postgres.authentication.username":        []byte(connectionPostgresAuthenticationUsername),
				"connection.postgres.certificate.certificate_base64": []byte(connectionPostgresCertificateCertificateBase64),
				"connection.postgres.certificate.name":               []byte(connectionPostgresCertificateName),
				"connection.postgres.composed.0":                     []byte(connectionPostgresComposed0),
				"connection.postgres.database":                       []byte(connectionPostgresDatabase),
				"connection.postgres.hosts.0.hostname":               []byte(connectionPostgresHosts0Hostname),
				"connection.postgres.hosts.0.port":                   []byte(connectionPostgresHosts0Port),
				"connection.postgres.path":                           []byte(connectionPostgresPath),
				"connection.postgres.query_options.sslmode":          []byte(connectionPostgresQueryOptionsSslmode),
				"connection.postgres.scheme":                         []byte(connectionPostgresScheme),
				"connection.postgres.type":                           []byte(connectionPostgresType),
				"instance_administration_api.deployment_id":          []byte(instanceAdministrationAPIDeploymentID),
				"instance_administration_api.instance_id":            []byte(instanceAdministrationAPIInstanceID),
				"instance_administration_api.root":                   []byte(instanceAdministrationAPIRoot),
			}},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			conn, err := GetConnectionDetails(tc.args.cr, tc.args.instance)
			if err != nil {
				t.Errorf("GetConnectionDetails(...): error %s\n", err)
			}
			if diff := cmp.Diff(tc.want.conn, conn); diff != "" {
				t.Errorf("GetConnectionDetails(...): -want, +got:\n%s", diff)
			}
		})
	}
}
