package kaffine

import (
	"sort"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type KRMFunctionCatalog struct {
	// required
	APIVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
	Spec       struct {
		KrmFunctions []*KRMFunctionDefinitionSpec `json:"krmFunctions"`
	} `json:"spec"`
	// optional
	Metadata *v1.ObjectMeta `json:"metadata,omitempty"`
}

func NewKRMFunctionCatalog(name string) (*KRMFunctionCatalog, error) {
	cat := &KRMFunctionCatalog{}
	cat.APIVersion = "config.kubernetes.io/v1alpha1"
	cat.Kind = "KRMFunctionCatalog"
	cat.Metadata = &v1.ObjectMeta{}
	cat.Metadata.Name = name
	cat.Metadata.SetCreationTimestamp(v1.Now())

	return cat, nil
}

func (m *KRMFunctionCatalog) Validate() error {
	return nil
}

type KRMFunctionDefinitionSpec struct {
	// required
	Group       string `json:"group"`
	Description string `json:"description"`
	Publisher   string `json:"publisher"`
	Names       struct {
		Kind string `json:"kind"`
	} `json:"names"`
	Versions []*KRMFunctionVersion `json:"versions"`
	// optional
	Home        string   `json:"home,omitempty"`
	Maintainers []string `json:"maintainers,omitempty"`
	Tags        []string `json:"tags,omitempty"`
}

func (m *KRMFunctionDefinitionSpec) Validate() error {
	return nil
}

// Right now just lexicographically compares the version names
func (m *KRMFunctionDefinitionSpec) GetHighestVersion() *KRMFunctionVersion {
	sort.Slice(m.Versions, func(i, j int) bool {
		return m.Versions[i].Name < m.Versions[j].Name
	})

	return m.Versions[len(m.Versions)-1]
}

func (m *KRMFunctionDefinitionSpec) GetShortName() string {
	return m.Group + "/" + m.Names.Kind + ":" + m.GetHighestVersion().Name
}

type KRMFunctionVersion struct {
	// required
	// Schema     struct{ OpenAPIV3Schema v1beta1.JSONSchemaProps `json:"openAPIV3Schema"` }  `json:"schema"`
	Name       string   `json:"name"`
	Idempotent bool     `json:"idempotent"`
	Usage      string   `json:"usage"`
	Examples   []string `json:"examples"`
	License    string   `json:"license"`
	Runtime    struct {
		Container *KRMFunctionRuntimeContainer `json:"container,omitempty"`
		Exec      *KRMFunctionRuntimeExec      `json:"exec,omitempty"`
	} `json:"runtime"`
	// optional
	Maintainers []string `json:"maintainers,omitempty"`
}

func (m *KRMFunctionVersion) Validate() error {
	return nil
}

type KRMFunctionRuntimeContainer struct {
	// required
	Image string `json:"image"`
	// optional
	Sha256              string `json:"sha256,omitempty"`
	RequireNetwork      bool   `json:"requireNetwork,omitempty"`
	RequireStorageMount bool   `json:"requireStorageMount,omitempty"`
}

func (m *KRMFunctionRuntimeContainer) Validate() error {
	return nil
}

type KRMFunctionRuntimeExec struct {
	// required
	Platforms []*KRMFunctionRuntimePlatform `json:"platforms"`
}

func (m *KRMFunctionRuntimeExec) Validate() error {
	return nil
}

type KRMFunctionRuntimePlatform struct {
	// required
	Bin    string `json:"bin"`
	Os     string `json:"os"`
	Arch   string `json:"arch"`
	Uri    string `json:"uri"`
	Sha256 string `json:"sha256"`
}

func (m *KRMFunctionRuntimePlatform) Validate() error {
	return nil
}
