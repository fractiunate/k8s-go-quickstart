package main

import (
	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/core/v1"
	helmv3 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/helm/v3"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		// isMinikube := config.GetBool(ctx, "isMinikube")
		// appName := "nginx"
		// appLabels := pulumi.StringMap{
		// 	"app": pulumi.String(appName),
		// }
		// deployment, err := appsv1.NewDeployment(ctx, appName, &appsv1.DeploymentArgs{
		// 	Spec: appsv1.DeploymentSpecArgs{
		// 		Selector: &metav1.LabelSelectorArgs{
		// 			MatchLabels: appLabels,
		// 		},
		// 		Replicas: pulumi.Int(1),
		// 		Template: &corev1.PodTemplateSpecArgs{
		// 			Metadata: &metav1.ObjectMetaArgs{
		// 				Labels: appLabels,
		// 			},
		// 			Spec: &corev1.PodSpecArgs{
		// 				Containers: corev1.ContainerArray{
		// 					corev1.ContainerArgs{
		// 						Name:  pulumi.String("nginx"),
		// 						Image: pulumi.String("nginx"),
		// 					}},
		// 			},
		// 		},
		// 	},
		// })
		// if err != nil {
		// 	return err
		// }

		// feType := "LoadBalancer"
		// if isMinikube {
		// 	feType = "ClusterIP"
		// }

		// template := deployment.Spec.ApplyT(func(v *appsv1.DeploymentSpec) *corev1.PodTemplateSpec {
		// 	return &v.Template
		// }).(corev1.PodTemplateSpecPtrOutput)

		// meta := template.ApplyT(func(v *corev1.PodTemplateSpec) *metav1.ObjectMeta { return v.Metadata }).(metav1.ObjectMetaPtrOutput)

		// frontend, err := corev1.NewService(ctx, appName, &corev1.ServiceArgs{
		// 	Metadata: meta,
		// 	Spec: &corev1.ServiceSpecArgs{
		// 		Type: pulumi.String(feType),
		// 		Ports: &corev1.ServicePortArray{
		// 			&corev1.ServicePortArgs{
		// 				Port:       pulumi.Int(80),
		// 				TargetPort: pulumi.Int(80),
		// 				Protocol:   pulumi.String("TCP"),
		// 			},
		// 		},
		// 		Selector: appLabels,
		// 	},
		// })
		// if err != nil {
		// 	return err
		// }

		// var ip pulumi.StringOutput

		// if isMinikube {
		// 	ip = frontend.Spec.ApplyT(func(val *corev1.ServiceSpec) string {
		// 		if val.ClusterIP != nil {
		// 			return *val.ClusterIP
		// 		}
		// 		return ""
		// 	}).(pulumi.StringOutput)
		// } else {
		// 	ip = frontend.Status.ApplyT(func(val *corev1.ServiceStatus) string {
		// 		if val.LoadBalancer.Ingress[0].Ip != nil {
		// 			return *val.LoadBalancer.Ingress[0].Ip
		// 		}
		// 		return *val.LoadBalancer.Ingress[0].Hostname
		// 	}).(pulumi.StringOutput)
		// }

		// func NewPersistentVolumeClaim(ctx *Context, name string, args *PersistentVolumeClaimArgs, opts ...ResourceOption) (*PersistentVolumeClaim, error)
		// pvc := corev1.NewPersistentVolumeClaim(ctx, "postgresvol")

		initSqlConfigMap, err := corev1.NewConfigMap(ctx, "init-sql-config", &corev1.ConfigMapArgs{
			Metadata: &metav1.ObjectMetaArgs{
				Namespace: pulumi.String("default"),
			},
			Data: pulumi.StringMap{
				"config": pulumi.String(`init-script.sh":"CREATE USER helm;CREATE DATABASE helm;GRANT ALL PRIVILEGES ON DATABASE helm TO helm;`),
			},
		})
		if err != nil {
			return err
		}

		// volumes: [{ name: "nginx-configs", configMap: { name: nginxConfigName } }],

		_, postgresqlErr := helmv3.NewChart(ctx, "bitnami-postgresql", helmv3.ChartArgs{
			Chart:   pulumi.String("postgresql"),
			Version: pulumi.String("10.16.1"),
			FetchArgs: helmv3.FetchArgs{
				Repo: pulumi.String("https://charts.bitnami.com/bitnami"),
			},
			Values: pulumi.Map{
				"initdbScriptsConfigMap": initSqlConfigMap.Metadata.Name(),
				"global": pulumi.Map{
					"postgresql": pulumi.Map{
						"postgresqlDatabase ": pulumi.String("pulumidb"),
						"postgresqlUsername":  pulumi.String("postgres"),
						"postgresqlPassword":  pulumi.String("postgres"),
					},
				},
			},
		})
		if postgresqlErr != nil {
			return postgresqlErr
		}

		_, pgadminErr := helmv3.NewChart(ctx, "pgadmin", helmv3.ChartArgs{
			Chart:   pulumi.String("pgadmin4"),
			Version: pulumi.String("1.9.0"),
			FetchArgs: helmv3.FetchArgs{
				Repo: pulumi.String("https://helm.runix.net/"),
			},
			Values: pulumi.Map{
				"env": pulumi.Map{
					"email":    pulumi.String("admin@test.de"),
					"password": pulumi.String("admin"),
				},
			},
		})
		if pgadminErr != nil {
			return pgadminErr
		}

		// nginx = Chart(
		// 	"nginx",
		// 	ChartOpts(
		// 		chart="nginx",
		// 		version="8.4.0",
		// 		namespace="app",
		// 		fetch_opts=FetchOpts(
		// 			repo="https://charts.bitnami.com/bitnami",
		// 		)
		// 	),
		// )

		// ctx.Export("ip", ip)
		return nil
	})
}
