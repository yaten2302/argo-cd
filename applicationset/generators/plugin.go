package generators

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/jeremywohl/flatten"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	argoprojiov1alpha1 "github.com/argoproj/argo-cd/v3/pkg/apis/application/v1alpha1"
	"github.com/argoproj/argo-cd/v3/util/settings"

	"github.com/argoproj/argo-cd/v3/applicationset/services/plugin"
)

const (
	DefaultPluginRequeueAfter = 30 * time.Minute
)

var _ Generator = (*PluginGenerator)(nil)

type PluginGenerator struct {
	client    client.Client
	namespace string
}

func NewPluginGenerator(client client.Client, namespace string) Generator {
	g := &PluginGenerator{
		client:    client,
		namespace: namespace,
	}
	return g
}

func (g *PluginGenerator) GetRequeueAfter(appSetGenerator *argoprojiov1alpha1.ApplicationSetGenerator) time.Duration {
	// Return a requeue default of 30 minutes, if no default is specified.

	if appSetGenerator.Plugin.RequeueAfterSeconds != nil {
		return time.Duration(*appSetGenerator.Plugin.RequeueAfterSeconds) * time.Second
	}

	return DefaultPluginRequeueAfter
}

func (g *PluginGenerator) GetTemplate(appSetGenerator *argoprojiov1alpha1.ApplicationSetGenerator) *argoprojiov1alpha1.ApplicationSetTemplate {
	return &appSetGenerator.Plugin.Template
}

func (g *PluginGenerator) GenerateParams(appSetGenerator *argoprojiov1alpha1.ApplicationSetGenerator, applicationSetInfo *argoprojiov1alpha1.ApplicationSet, _ client.Client) ([]map[string]any, error) {
	if appSetGenerator == nil {
		return nil, ErrEmptyAppSetGenerator
	}

	if appSetGenerator.Plugin == nil {
		return nil, ErrEmptyAppSetGenerator
	}

	ctx := context.Background()

	providerConfig := appSetGenerator.Plugin

	pluginClient, err := g.getPluginFromGenerator(ctx, applicationSetInfo.Name, providerConfig)
	if err != nil {
		return nil, fmt.Errorf("error getting plugin from generator: %w", err)
	}

	list, err := pluginClient.List(ctx, providerConfig.Input.Parameters)
	if err != nil {
		return nil, fmt.Errorf("error listing params: %w", err)
	}

	res, err := g.generateParams(appSetGenerator, applicationSetInfo, list.Output.Parameters, appSetGenerator.Plugin.Input.Parameters, applicationSetInfo.Spec.GoTemplate)
	if err != nil {
		return nil, fmt.Errorf("error generating params: %w", err)
	}

	return res, nil
}

func (g *PluginGenerator) getPluginFromGenerator(ctx context.Context, appSetName string, generatorConfig *argoprojiov1alpha1.PluginGenerator) (*plugin.Service, error) {
	cm, err := g.getConfigMap(ctx, generatorConfig.ConfigMapRef.Name)
	if err != nil {
		return nil, fmt.Errorf("error fetching ConfigMap: %w", err)
	}
	token, err := g.getToken(ctx, cm["token"])
	if err != nil {
		return nil, fmt.Errorf("error fetching Secret token: %w", err)
	}

	var requestTimeout int
	requestTimeoutStr, ok := cm["requestTimeout"]
	if ok {
		requestTimeout, err = strconv.Atoi(requestTimeoutStr)
		if err != nil {
			return nil, fmt.Errorf("error set requestTimeout : %w", err)
		}
	}

	pluginClient, err := plugin.NewPluginService(appSetName, cm["baseUrl"], token, requestTimeout)
	if err != nil {
		return nil, fmt.Errorf("error initializing plugin client: %w", err)
	}
	return pluginClient, nil
}

func (g *PluginGenerator) generateParams(appSetGenerator *argoprojiov1alpha1.ApplicationSetGenerator, appSet *argoprojiov1alpha1.ApplicationSet, objectsFound []map[string]any, pluginParams argoprojiov1alpha1.PluginParameters, useGoTemplate bool) ([]map[string]any, error) {
	res := []map[string]any{}

	for _, objectFound := range objectsFound {
		params := map[string]any{}

		if useGoTemplate {
			for k, v := range objectFound {
				params[k] = v
			}
		} else {
			flat, err := flatten.Flatten(objectFound, "", flatten.DotStyle)
			if err != nil {
				return nil, err
			}
			for k, v := range flat {
				params[k] = fmt.Sprintf("%v", v)
			}
		}

		params["generator"] = map[string]any{
			"input": map[string]argoprojiov1alpha1.PluginParameters{
				"parameters": pluginParams,
			},
		}

		err := appendTemplatedValues(appSetGenerator.Plugin.Values, params, appSet.Spec.GoTemplate, appSet.Spec.GoTemplateOptions)
		if err != nil {
			return nil, err
		}

		res = append(res, params)
	}

	return res, nil
}

func (g *PluginGenerator) getToken(ctx context.Context, tokenRef string) (string, error) {
	if tokenRef == "" || !strings.HasPrefix(tokenRef, "$") {
		return "", fmt.Errorf("token is empty, or does not reference a secret key starting with '$': %v", tokenRef)
	}

	secretName, tokenKey := plugin.ParseSecretKey(tokenRef)

	secret := &corev1.Secret{}
	err := g.client.Get(
		ctx,
		client.ObjectKey{
			Name:      secretName,
			Namespace: g.namespace,
		},
		secret)
	if err != nil {
		return "", fmt.Errorf("error fetching secret %s/%s: %w", g.namespace, secretName, err)
	}

	secretValues := make(map[string]string, len(secret.Data))

	for k, v := range secret.Data {
		secretValues[k] = string(v)
	}

	token := settings.ReplaceStringSecret(tokenKey, secretValues)

	return token, err
}

func (g *PluginGenerator) getConfigMap(ctx context.Context, configMapRef string) (map[string]string, error) {
	cm := &corev1.ConfigMap{}
	err := g.client.Get(
		ctx,
		client.ObjectKey{
			Name:      configMapRef,
			Namespace: g.namespace,
		},
		cm)
	if err != nil {
		return nil, err
	}

	baseURL, ok := cm.Data["baseUrl"]
	if !ok || baseURL == "" {
		return nil, errors.New("baseUrl not found in ConfigMap")
	}

	token, ok := cm.Data["token"]
	if !ok || token == "" {
		return nil, errors.New("token not found in ConfigMap")
	}

	return cm.Data, nil
}
