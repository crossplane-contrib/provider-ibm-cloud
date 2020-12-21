module github.com/crossplane-contrib/provider-ibm-cloud

go 1.13

require (
	github.com/IBM-Cloud/bluemix-go v0.0.0-20201019071904-51caa09553fb
	github.com/IBM/experimental-go-sdk v0.0.0-20201217171549-fc1a6f324de2
	github.com/IBM/go-sdk-core v1.1.0
	github.com/IBM/platform-services-go-sdk v0.14.4
	github.com/crossplane/crossplane-runtime v0.11.1-0.20201116232334-1b691efff491
	github.com/crossplane/crossplane-tools v0.0.0-20201007233256-88b291e145bb
	github.com/go-openapi/strfmt v0.19.11
	github.com/google/go-cmp v0.5.2
	github.com/jeremywohl/flatten v1.0.1
	github.com/pkg/errors v0.9.1
	github.ibm.com/ibmcloud/databases-go-sdk v0.0.0-20201125155449-6e36d5cb805f // indirect
	gopkg.in/alecthomas/kingpin.v2 v2.2.6
	k8s.io/api v0.18.8
	k8s.io/apimachinery v0.18.8
	k8s.io/client-go v0.18.8
	sigs.k8s.io/controller-runtime v0.6.2
	sigs.k8s.io/controller-tools v0.2.4
)
