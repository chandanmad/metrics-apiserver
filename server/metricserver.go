package server

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/apimachinery/announced"
	"k8s.io/apimachinery/pkg/apimachinery/registered"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	genericapiserver "k8s.io/apiserver/pkg/server"
	cm "k8s.io/metrics/pkg/apis/custom_metrics"
	cminstall "k8s.io/metrics/pkg/apis/custom_metrics/install"
	_ "k8s.io/metrics/pkg/apis/custom_metrics/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/version"
	"k8s.io/apiserver/pkg/registry/rest"
	genericapirequest "k8s.io/apiserver/pkg/endpoints/request"
	"k8s.io/apimachinery/pkg/util/wait"
	"fmt"
	"net"
	"k8s.io/apiserver/pkg/server/options"
	"k8s.io/kubernetes/staging/src/k8s.io/apiserver/pkg/endpoints/request"
	"github.com/pkg/errors"
)

var (
	groupFactoryRegistry = make(announced.APIGroupFactoryRegistry)
	registry             = registered.NewOrDie("")
	scheme               = runtime.NewScheme()
	codecs               = serializer.NewCodecFactory(scheme)
	groupResources       = []schema.GroupResource{}
)

func init() {
	cminstall.Install(groupFactoryRegistry, registry, scheme)
	metav1.AddToGroupVersion(scheme, schema.GroupVersion{Version: "v1"})
	scheme.AddUnversionedTypes(schema.GroupVersion{Group: "", Version: "v1"},
		&metav1.Status{},
		&metav1.APIVersions{},
		&metav1.APIGroupList{},
		&metav1.APIGroup{},
		&metav1.APIResourceList{},)
}

type MetricServer struct {
	groupResources []schema.GroupResource
	genericApiServer *genericapiserver.GenericAPIServer
}


func NewMetricServer() (*MetricServer, error) {
	config := genericapiserver.NewConfig(codecs)
	secureServingOptions := options.NewSecureServingOptions()
	secureServingOptions.ForceLoopbackConfigUsage()
	if err := secureServingOptions.MaybeDefaultWithSelfSignedCerts("localhost", nil, []net.IP{net.ParseIP("127.0.0.1")}); err != nil {
		return nil, err
	}
	secureServingOptions.ApplyTo(config)
	config.ReadWritePort = 443
	config.Version = &version.Info{Major: "1", Minor: "0"}

	genericserver, err := config.SkipComplete().New("metrics-apiserver", genericapiserver.EmptyDelegate)
	if err != nil {
		return nil, err
	}
	apiGroupInfo := genericapiserver.NewDefaultAPIGroupInfo(cm.GroupName, registry, scheme, metav1.ParameterCodec, codecs)

	storage := &MetricSource{}
	cmStorage := map[string]rest.Storage{}
	cmStorage["pods/{subresource}"] = storage
	cmStorage["pods"] = storage
	apiGroupInfo.VersionedResourcesStorageMap["v1alpha1"] = cmStorage

	if err := genericserver.InstallAPIGroup(&apiGroupInfo); err != nil {
		return nil, err
	}
	fmt.Printf("server config %+v", config)
	return &MetricServer{groupResources: []schema.GroupResource{}, genericApiServer: genericserver}, nil
}

func (metricServer *MetricServer) Start() error {
	if err := metricServer.genericApiServer.PrepareRun().Run(wait.NeverStop); err != nil {
		return err
	}
	return nil
}

// implements storage and getter interfaces
type MetricSource struct {
	name string
}

func (source *MetricSource) New() runtime.Object {
	return &cm.MetricValueList{}
}

func (source *MetricSource) Get(ctx genericapirequest.Context, name string, options *metav1.GetOptions) (runtime.Object, error) {
	_, ok := request.RequestInfoFrom(ctx)
	if !ok {
		return nil, errors.New("RequestInfo not found")
	}
	metricValueList := &cm.MetricValueList{}
	fmt.Println("getter called")
	return metricValueList, nil
}



func (server *MetricServer) AddGroupResource(group , resource string) {
	groupResources = append(groupResources, schema.GroupResource{Group: group, Resource: resource})
}




