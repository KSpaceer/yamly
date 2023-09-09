package benchmark

type ExtraLargeStruct struct {
	Data []LargeStruct
}

type LargeStruct = DeploymentManifest

type SmallStruct struct {
	Name      string   `yaml:"name"`
	Age       int      `yaml:"age"`
	Nicknames []string `yaml:"nicknames"`
}

type DeploymentManifest struct {
	APIVersion string         `yaml:"apiVersion,omitempty,omitempty"`
	Kind       string         `yaml:"kind,omitempty,omitempty"`
	Metadata   ObjectMeta     `yaml:"metadata,omitempty,omitempty"`
	Spec       DeploymentSpec `yaml:"spec,omitempty,omitempty"`
}

type ObjectMeta struct {
	Name        string            `yaml:"name,omitempty"`
	Namespace   string            `yaml:"namespace,omitempty"`
	Labels      map[string]string `yaml:"labels,omitempty"`
	Annotations map[string]string `yaml:"annotations,omitempty"`
}

type DeploymentSpec struct {
	Replicas                int `yaml:"replicas,omitempty"`
	MinReadySeconds         int `yaml:"minReadySeconds,omitempty"`
	ProgressDeadlineSeconds int `yaml:"progressDeadlineSeconds,omitempty"`
	RevisionHistoryLimit    int `yaml:"revisionHistoryLimit,omitempty"`

	Strategy DeploymentStrategy `yaml:"strategy,omitempty"`

	Selector LabelSelector `yaml:"selector,omitempty"`

	Template PodTemplateSpec `yaml:"template,omitempty"`
}

type DeploymentStrategy struct {
	RollingUpdate RollingUpdateDeployment `yaml:"rollingUpdate,omitempty"`
	Type          string                  `yaml:"type,omitempty"`
}

type RollingUpdateDeployment struct {
	MaxSurge       int `yaml:"maxSurge,omitempty"`
	MaxUnavailable int `yaml:"maxUnavailable,omitempty"`
}

type LabelSelector struct {
	MatchLabels map[string]string `yaml:"matchLabels,omitempty"`
}

type PodTemplateSpec struct {
	Metadata ObjectMeta `yaml:"metadata,omitempty"`
	Spec     PodSpec    `yaml:"spec,omitempty"`
}

type PodSpec struct {
	RestartPolicy    string                 `yaml:"restartPolicy,omitempty"`
	SecurityContext  PodSecurityContext     `yaml:"securityContext,omitempty"`
	HostAliases      []HostAlias            `yaml:"hostAliases,omitempty"`
	NodeSelector     map[string]string      `yaml:"nodeSelector,omitempty"`
	ImagePullSecrets []LocalObjectReference `yaml:"imagePullSecrets,omitempty"`
	Containers       []Container            `yaml:"containers,omitempty"`
	InitContainers   []Container            `yaml:"initContainers,omitempty"`
	Volumes          []Volume               `yaml:"volumes,omitempty"`
}

type PodSecurityContext struct {
	RunAsUser    int  `yaml:"runAsUser,omitempty"`
	RunAsGroup   int  `yaml:"runAsGroup,omitempty"`
	RunAsNonRoot bool `yaml:"runAsNonRoot,omitempty"`
}

type HostAlias struct {
	IP        string   `yaml:"ip"`
	Hostnames []string `yaml:"hostnames"`
}

type LocalObjectReference struct {
	Name string `yaml:"name"`
}

type Container struct {
	Name            string               `yaml:"name"`
	Image           string               `yaml:"image"`
	ImagePullPolicy string               `yaml:"imagePullPolicy,omitempty"`
	Command         []string             `yaml:"command,omitempty"`
	Ports           []ContainerPort      `yaml:"ports,omitempty"`
	Env             []EnvVar             `yaml:"env,omitempty"`
	EnvFrom         []EnvFromSource      `yaml:"envFrom,omitempty"`
	Resources       ResourceRequirements `yaml:"resources,omitempty"`
	LivenessProbe   Probe                `yaml:"livenessProbe,omitempty"`
	ReadinessProbe  Probe                `yaml:"readinessProbe,omitempty"`
	StartupProbe    Probe                `yaml:"startupProbe,omitempty"`
	Lifecycle       Lifecycle            `yaml:"lifecycle,omitempty"`
	VolumeMounts    []VolumeMount        `yaml:"volumeMounts,omitempty"`
	SecurityContext SecurityContext      `yaml:"securityContext,omitempty"`
}

type ContainerPort struct {
	ContainerPort int `yaml:"containerPort,omitempty"`
}

type EnvVar struct {
	Name      string       `yaml:"name"`
	Value     string       `yaml:"value,omitempty"`
	ValueFrom EnvVarSource `yaml:"valueFrom,omitempty"`
}

type EnvVarSource struct {
	SecretKeyRef SecretKeySelector   `yaml:"secretKeyRef,omitempty"`
	FieldRef     ObjectFieldSelector `yaml:"fieldRef,omitempty"`
}

type SecretKeySelector struct {
	Key      string `yaml:"key"`
	Name     string `yaml:"name"`
	Optional bool   `yaml:"optional,omitempty"`
}

type ObjectFieldSelector struct {
	APIVersion string `yaml:"apiVersion,omitempty"`
	FieldPath  string `yaml:"fieldPath,omitempty"`
}

type EnvFromSource struct {
	ConfigMapRef ConfigMapEnvSource `yaml:"configMapRef,omitempty"`
	SecretRef    SecretEnvSource    `yaml:"secretRef,omitempty"`
}

type ConfigMapEnvSource struct {
	Name     string `yaml:"name"`
	Optional bool   `yaml:"optional,omitempty"`
}

type SecretEnvSource struct {
	Name     string `yaml:"name"`
	Optional bool   `yaml:"optional,omitempty"`
}

type ResourceRequirements struct {
	Requests map[string]string `yaml:"requests,omitempty"`
	Limits   map[string]string `yaml:"limits,omitempty"`
}

type Probe struct {
	Exec                ExecAction    `yaml:"exec,omitempty"`
	FailureThreshold    int           `yaml:"failureThreshold,omitempty"`
	HTTPGet             HTTPGetAction `yaml:"httpGet,omitempty"`
	InitialDelaySeconds int           `yaml:"initialDelaySeconds,omitempty"`
	PeriodSeconds       int           `yaml:"periodSeconds,omitempty"`
}

type ExecAction struct {
	Command []string `yaml:"command"`
}

type HTTPGetAction struct {
	Host        string       `yaml:"host,omitempty"`
	HTTPHeaders []HTTPHeader `yaml:"httpHeaders,omitempty"`
	Path        string       `yaml:"path"`
	Port        string       `yaml:"port,omitempty"`
	Scheme      string       `yaml:"scheme,omitempty"`
}

type HTTPHeader struct {
	Name  string `yaml:"name"`
	Value string `yaml:"value"`
}

type Lifecycle struct {
	PostStart LifecycleHandler `yaml:"postStart,omitempty"`
	PreStop   LifecycleHandler `yaml:"preStop,omitempty"`
}

type LifecycleHandler struct {
	Exec    ExecAction    `yaml:"exec,omitempty"`
	HTTPGet HTTPGetAction `yaml:"httpGet,omitempty"`
}

type VolumeMount struct {
	MountPath string `yaml:"mountPath"`
	Name      string `yaml:"name"`
	ReadOnly  bool   `yaml:"readOnly,omitempty"`
}

type SecurityContext struct {
	Privileged   bool `yaml:"privileged,omitempty"`
	RunAsUser    int  `yaml:"runAsUser,omitempty"`
	RunAsGroup   int  `yaml:"runAsGroup,omitempty"`
	RunAsNonRoot bool `yaml:"runAsNonRoot,omitempty"`
}

type Volume struct {
	Name                  string                            `yaml:"name"`
	EmptyDir              EmptyDirVolumeSource              `yaml:"emptyDir,omitempty"`
	PersistentVolumeClaim PersistentVolumeClaimVolumeSource `yaml:"persistentVolumeClaim,omitempty"`
	NFS                   NFSVolumeSource                   `yaml:"nfs,omitempty"`
	ConfigMap             ConfigMapVolumeSource             `yaml:"configMap,omitempty"`
}

type EmptyDirVolumeSource struct {
	Medium    string `yaml:"medium,omitempty"`
	SizeLimit string `yaml:"sizeLimit,omitempty"`
}

type PersistentVolumeClaimVolumeSource struct {
	ClaimName string `yaml:"claimName"`
	ReadOnly  bool   `yaml:"readOnly,omitempty"`
}

type NFSVolumeSource struct {
	Server   string `yaml:"server"`
	Path     string `yaml:"path"`
	ReadOnly bool   `yaml:"readOnly,omitempty"`
}

type ConfigMapVolumeSource struct {
	DefaultMode int         `yaml:"defaultMode,omitempty"`
	Name        string      `yaml:"name"`
	Items       []KeyToPath `yaml:"items"`
	Optional    bool        `yaml:"optional,omitempty"`
}

type KeyToPath struct {
	Key  string `yaml:"key"`
	Path string `yaml:"path"`
	Mode int    `yaml:"mode,omitempty"`
}
