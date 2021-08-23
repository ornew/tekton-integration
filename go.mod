module github.com/ornew/tekton-integration

go 1.16

require (
	github.com/bradleyfalzon/ghinstallation v1.1.1
	github.com/cloudevents/sdk-go/v2 v2.5.0
	github.com/coocood/freecache v1.1.1
	github.com/go-logr/logr v0.4.0
	github.com/google/go-github/v37 v37.0.0
	github.com/google/uuid v1.2.0
	github.com/onsi/ginkgo v1.16.4
	github.com/onsi/gomega v1.13.0
	github.com/stretchr/testify v1.7.0
	github.com/tektoncd/pipeline v0.26.0
	k8s.io/api v0.21.2
	k8s.io/apimachinery v0.21.2
	k8s.io/client-go v0.21.2
	k8s.io/utils v0.0.0-20210527160623-6fdb442a123b
	knative.dev/pkg v0.0.0-20210510175900-4564797bf3b7
	sigs.k8s.io/controller-runtime v0.9.2
)
