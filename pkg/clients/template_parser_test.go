package clients

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestTemplateParser(t *testing.T) {
	type args struct {
		vars map[string]string
		obj  map[string]interface{}
	}
	type want struct {
		values map[string]interface{}
	}
	cases := map[string]struct {
		args args
		want want
	}{
		"Simple": {
			args: args{
				vars: map[string]string{
					"apikey":    `{{ .apikey }}`,
					"endpoints": `{{ .endpoints }}`,
				},
				obj: map[string]interface{}{
					"apikey":                 "fake-key",
					"endpoints":              "https://control.cloud-object-storage.cloud.ibm.com/v2/endpoints",
					"iam_apikey_description": "Auto-generated for key xyz",
					"iam_apikey_name":        "Service credentials-1",
					"iam_role_crn":           "crn:v1:bluemix:public:iam::::serviceRole:Manager",
					"iam_serviceid_crn":      "crn:v1:bluemix:public:iam-identity::a/x::serviceid:ServiceId-z",
					"resource_instance_id":   "crn:v1:bluemix:public:cloud-object-storage:global:a/x:y::",
				},
			},
			want: want{values: map[string]interface{}{
				"apikey":    "fake-key",
				"endpoints": "https://control.cloud-object-storage.cloud.ibm.com/v2/endpoints",
			},
			},
		},
		"Complex": {
			args: args{
				vars: map[string]string{
					"JDBC_CONNECTION_STRING": `jdbc:postgresql://{{ (index .connection.postgres.hosts 0).hostname }}:{{ (index .connection.postgres.hosts 0).port }}{{ .connection.postgres.path }}`,
					"USER":                   `{{ .connection.postgres.authentication.username }}`,
					"PASSWORD":               `{{ .connection.postgres.authentication.password }}`,
				},
				obj: map[string]interface{}{
					"connection": map[string]interface{}{
						"postgres": map[string]interface{}{
							"authentication": map[string]interface{}{
								"method":   "direct",
								"username": "database-user",
								"password": "database-pass",
							},
							"hosts": []map[string]interface{}{
								{
									"hostname": "host1.databases.appdomain.cloud",
									"port":     31700,
								},
							},
							"path": "/ibmclouddb",
						},
					},
				},
			},
			want: want{values: map[string]interface{}{
				"PASSWORD":               "database-pass",
				"JDBC_CONNECTION_STRING": "jdbc:postgresql://host1.databases.appdomain.cloud:31700/ibmclouddb",
				"USER":                   "database-user",
			},
			},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			parser := NewTemplateParser(tc.args.vars, tc.args.obj)
			values, err := parser.Parse()
			if err != nil {
				t.Errorf("TestTemplateParser(...): -want, +got:\n%s", err)
			}
			if diff := cmp.Diff(tc.want.values, values); diff != "" {
				t.Errorf("TestTemplateParser(...): -want, +got:\n%s", diff)
			}

		})
	}
}
