package benchmark

import _ "embed"

//go:embed example.yaml
var largeDataText []byte

var largeData = DeploymentManifest{
	APIVersion: "apps/v1",
	Kind:       "Deployment",
	Metadata: ObjectMeta{
		Name:      "website",
		Namespace: "default",
		Labels: map[string]string{
			"app": "website",
		},
		Annotations: map[string]string{
			"source_url": "git@gitlab.com:kisphp/example.git",
		},
	},
	Spec: DeploymentSpec{
		Replicas:                1,
		MinReadySeconds:         10,
		ProgressDeadlineSeconds: 60,
		RevisionHistoryLimit:    5,
		Strategy: DeploymentStrategy{
			RollingUpdate: RollingUpdateDeployment{
				MaxSurge:       1,
				MaxUnavailable: 0,
			},
		},
		Selector: LabelSelector{
			MatchLabels: map[string]string{
				"app": "website",
			},
		},
		Template: PodTemplateSpec{
			Metadata: ObjectMeta{
				Name: "website",
				Labels: map[string]string{
					"app": "website",
				},
				Annotations: map[string]string{
					"source_url": "git@gitlab.com:kisphp/example.git",
				},
			},
			Spec: PodSpec{
				RestartPolicy: "Always",
				SecurityContext: PodSecurityContext{
					RunAsUser:    33,
					RunAsGroup:   33,
					RunAsNonRoot: true,
				},
				HostAliases: []HostAlias{
					{
						IP: "127.0.0.1",
						Hostnames: []string{
							"foo.local",
							"bar.local",
						},
					},
					{
						IP: "10.1.2.3",
						Hostnames: []string{
							"foo.remote",
							"bar.remote",
						},
					},
				},
				NodeSelector: map[string]string{
					"type": "application",
				},
				ImagePullSecrets: []LocalObjectReference{
					{
						Name: "my-registry-secret",
					},
				},
				Containers: []Container{
					{
						Name:            "website",
						Image:           "nginx:latest",
						ImagePullPolicy: "Always",
						Ports: []ContainerPort{
							{
								ContainerPort: 80,
							},
						},
						Env: []EnvVar{
							{
								Name:  "APP_TYPE",
								Value: "application",
							},
							{
								Name: "APP_SECRET",
								ValueFrom: EnvVarSource{
									SecretKeyRef: SecretKeySelector{
										Key:  "APP_SECRET",
										Name: "db-secrets",
									},
								},
							},
							{
								Name: "K8S_NODE_NAME",
								ValueFrom: EnvVarSource{
									FieldRef: ObjectFieldSelector{
										FieldPath: "spec.nodeName",
									},
								},
							},
							{
								Name: "K8S_POD_NAME",
								ValueFrom: EnvVarSource{
									FieldRef: ObjectFieldSelector{
										FieldPath: "metadata.name",
									},
								},
							},
							{
								Name: "K8S_POD_NAMESPACE",
								ValueFrom: EnvVarSource{
									FieldRef: ObjectFieldSelector{
										FieldPath: "metadata.namespace",
									},
								},
							},
							{
								Name: "K8S_POD_IP",
								ValueFrom: EnvVarSource{
									FieldRef: ObjectFieldSelector{
										FieldPath: "status.podIP",
									},
								},
							},
							{
								Name: "K8S_POD_SERVICE_ACCOUNT",
								ValueFrom: EnvVarSource{
									FieldRef: ObjectFieldSelector{
										FieldPath: "spec.serviceAccountName",
									},
								},
							},
						},
						EnvFrom: []EnvFromSource{
							{
								ConfigMapRef: ConfigMapEnvSource{
									Name: "site-configurations",
								},
							},
							{
								SecretRef: SecretEnvSource{
									Name: "site-secrets",
								},
							},
						},
						Resources: ResourceRequirements{
							Requests: map[string]string{
								"memory": "64Mi",
								"cpu":    "10m",
							},
							Limits: map[string]string{
								"memory": "256Mi",
								"cpu":    "100m",
							},
						},
						LivenessProbe: Probe{
							HTTPGet: HTTPGetAction{
								Path: "/healthz",
								Port: "8080",
								HTTPHeaders: []HTTPHeader{
									{
										Name:  "Custom-Header",
										Value: "Awesome",
									},
								},
							},
							InitialDelaySeconds: 3,
							PeriodSeconds:       3,
						},
						ReadinessProbe: Probe{
							Exec: ExecAction{
								Command: []string{"cat", "/tmp/healthy"},
							},
							InitialDelaySeconds: 5,
							PeriodSeconds:       5,
						},
						StartupProbe: Probe{
							HTTPGet: HTTPGetAction{
								Path: "/healthz",
								Port: "liveness-port",
							},
							FailureThreshold: 30,
							PeriodSeconds:    10,
						},
						Lifecycle: Lifecycle{
							PostStart: LifecycleHandler{
								Exec: ExecAction{
									Command: []string{
										"/bin/bash",
										"-c",
										"curl -s -X GET --max-time 60 http://${SERVICE_NAME}.notifications.svc.cluster.local/start/${HOSTNAME}/php >&1; exit 0",
									},
								},
							},
							PreStop: LifecycleHandler{
								Exec: ExecAction{
									Command: []string{
										"/bin/bash",
										"-c",
										"curl -s -X GET --max-time 60 http://${SERVICE_NAME}.notifications.svc.cluster.local/stop/${HOSTNAME}/php >&1; exit 0",
									},
								},
							},
						},
						VolumeMounts: []VolumeMount{
							{
								MountPath: "/app/public/thumbs",
								Name:      "thumbnails",
							},
							{
								MountPath: "/app/uploads",
								Name:      "uploads",
							},
							{
								Name:      "config",
								MountPath: "/config",
								ReadOnly:  true,
							},
						},
					},
				},
				InitContainers: []Container{
					{
						Name:  "update-database",
						Image: "php-container",
						EnvFrom: []EnvFromSource{
							{
								ConfigMapRef: ConfigMapEnvSource{
									Name: "db-credentials",
								},
							},
						},
						Command: []string{
							"bin/console",
							"setup:install",
						},
						VolumeMounts: []VolumeMount{
							{
								MountPath: "/opt/test",
								Name:      "test",
							},
						},
						SecurityContext: SecurityContext{
							Privileged: true,
							RunAsUser:  0,
							RunAsGroup: 0,
						},
					},
				},
				Volumes: []Volume{
					{
						Name:     "thumbnails",
						EmptyDir: EmptyDirVolumeSource{},
					},
					{
						Name: "uploads",
						PersistentVolumeClaim: PersistentVolumeClaimVolumeSource{
							ClaimName: "website-uploads",
						},
					},
					{
						Name: "test",
						PersistentVolumeClaim: PersistentVolumeClaimVolumeSource{
							ClaimName: "my-test-volume",
						},
					},
					{
						Name: "efs-data",
						NFS: NFSVolumeSource{
							Server: "1a2b3c4d.efs.eu-central-1.amazonaws.com",
							Path:   "/",
						},
					},
					{
						Name: "config-volume",
						ConfigMap: ConfigMapVolumeSource{
							Name: "special-config",
							Items: []KeyToPath{
								{
									Key:  "SPECIAL_LEVEL",
									Path: "keys",
								},
							},
						},
					},
					{
						Name: "config",
						ConfigMap: ConfigMapVolumeSource{
							Name: "my-app-config",
							Items: []KeyToPath{
								{
									Key:  "game.properties",
									Path: "game.properties",
								},
								{
									Key:  "user-interface.properties",
									Path: "user-interface.properties",
								},
							},
						},
					},
				},
			},
		},
	},
}

var extraLargeData ExtraLargeStruct

func init() {
	for i := 0; i < 25; i++ {
		extraLargeData.Data = append(extraLargeData.Data, largeData)
	}
}

var smallData = SmallStruct{
	Name: "name",
	Age:  33,
	Nicknames: []string{
		"first",
		"second",
		"third",
	},
}

var smallDataText = []byte(`
name: name
age: 33
nicknames:
  - first
  - second
  - third
`)
