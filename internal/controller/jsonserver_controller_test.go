package controller

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	examplev1 "github.com/BlueTurtle-bytes/json-server/api/v1"
)

var _ = Describe("JsonServer Controller", func() {
	Context("When reconciling a JsonServer resource", func() {
		const resourceName = "app-test"

		ctx := context.Background()
		namespacedName := types.NamespacedName{
			Name:      resourceName,
			Namespace: "default",
		}

		BeforeEach(func() {
			By("Creating a JsonServer resource")
			js := &examplev1.JsonServer{}
			err := k8sClient.Get(ctx, namespacedName, js)

			if err != nil && errors.IsNotFound(err) {
				js = &examplev1.JsonServer{
					ObjectMeta: metav1.ObjectMeta{
						Name:      resourceName,
						Namespace: "default",
					},
					Spec: examplev1.JsonServerSpec{
						JsonConfig: `{
							"people": [
								{ "id": 1, "name": "Test" }
							]
						}`,
					},
				}
				Expect(k8sClient.Create(ctx, js)).To(Succeed())
			}
		})

		AfterEach(func() {
			By("Cleaning up the JsonServer resource")
			js := &examplev1.JsonServer{}
			err := k8sClient.Get(ctx, namespacedName, js)
			if err == nil {
				Expect(k8sClient.Delete(ctx, js)).To(Succeed())
			}
		})

		It("should create Deployment, Service, ConfigMap and update status", func() {
			By("Waiting for Deployment to be created")
			Eventually(func() error {
				deploy := &appsv1.Deployment{}
				return k8sClient.Get(ctx, namespacedName, deploy)
			}).Should(Succeed())

			By("Waiting for ConfigMap to be created")
			Eventually(func() error {
				cm := &corev1.ConfigMap{}
				return k8sClient.Get(ctx, namespacedName, cm)
			}).Should(Succeed())

			By("Waiting for Service to be created")
			Eventually(func() error {
				svc := &corev1.Service{}
				return k8sClient.Get(ctx, namespacedName, svc)
			}).Should(Succeed())

			By("Waiting for JsonServer status to be Synced")
			Eventually(func() string {
				js := &examplev1.JsonServer{}
				_ = k8sClient.Get(ctx, namespacedName, js)
				return js.Status.State
			}).Should(Equal("Synced"))
		})

	})
})
