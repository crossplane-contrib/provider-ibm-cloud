package resourcekey

import (
	"testing"

	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/google/go-cmp/cmp"

	rcv2 "github.com/IBM/platform-services-go-sdk/resourcecontrollerv2"

	"github.com/crossplane-contrib/provider-ibm-cloud/apis/resourcecontrollerv2/v1alpha1"
	ibmc "github.com/crossplane-contrib/provider-ibm-cloud/pkg/clients"
)

var (
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
				"iamRoleCrn":           []byte(iamRoleCRN),
				"iamServiceidCrn":      []byte(iamServiceidCRN),
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
				"iamRoleCrn":                   nil,
				"iamServiceidCrn":              nil,
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
