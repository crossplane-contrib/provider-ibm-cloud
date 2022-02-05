module github.com/crossplane-contrib/provider-ibm-cloud

go 1.13

require (
	github.com/IBM-Cloud/bluemix-go v0.0.0-20201019071904-51caa09553fb
	github.com/IBM/cloudant-go-sdk v0.0.34
	github.com/IBM/eventstreams-go-sdk v1.1.0
	github.com/IBM/experimental-go-sdk v0.0.0-20210112204617-192fc5b15655
	github.com/IBM/go-sdk-core v1.1.0
	github.com/IBM/go-sdk-core/v4 v4.10.0
	github.com/IBM/ibm-cos-sdk-go v1.7.0
	github.com/IBM/ibm-cos-sdk-go-config v1.2.0
	github.com/IBM/platform-services-go-sdk v0.17.18
	github.com/crossplane/crossplane-runtime v0.11.1-0.20201116232334-1b691efff491
	github.com/crossplane/crossplane-tools v0.0.0-20201007233256-88b291e145bb
	github.com/go-openapi/strfmt v0.21.1
	github.com/golang-jwt/jwt v3.2.2+incompatible // indirect
	github.com/google/go-cmp v0.5.5
	github.com/google/uuid v1.1.2 // indirect
	github.com/jeremywohl/flatten v1.0.1
	github.com/pkg/errors v0.9.1
	golang.org/x/oauth2 v0.0.0-20210514164344-f6687ab2804c // indirect
	golang.org/x/tools v0.1.7 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	gopkg.in/alecthomas/kingpin.v2 v2.2.6
	k8s.io/api v0.18.8
	k8s.io/apimachinery v0.18.8
	k8s.io/client-go v0.18.8
	sigs.k8s.io/controller-runtime v0.6.2
	sigs.k8s.io/controller-tools v0.2.4
)

replace github.com/dgrijalva/jwt-go => github.com/golang-jwt/jwt v3.2.1+incompatible
