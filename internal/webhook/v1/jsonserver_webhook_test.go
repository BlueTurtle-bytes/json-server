package v1

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	examplev1 "github.com/BlueTurtle-bytes/json-server/api/v1"
)

var _ = Describe("JsonServer Webhook", func() {
	var (
		ctx       context.Context
		validator JsonServerCustomValidator
	)

	BeforeEach(func() {
		ctx = context.Background()
		validator = JsonServerCustomValidator{}
	})

	Context("ValidateCreate", func() {
		It("should deny creation when name does not start with app-", func() {
			obj := &examplev1.JsonServer{
				ObjectMeta: metav1.ObjectMeta{
					Name: "invalid-name",
				},
				Spec: examplev1.JsonServerSpec{
					JsonConfig: `{}`,
				},
			}

			_, err := validator.ValidateCreate(ctx, obj)
			Expect(err).To(HaveOccurred())
		})

		It("should deny creation when jsonConfig is invalid JSON", func() {
			obj := &examplev1.JsonServer{
				ObjectMeta: metav1.ObjectMeta{
					Name: "app-valid",
				},
				Spec: examplev1.JsonServerSpec{
					JsonConfig: `{ invalid json }`,
				},
			}

			_, err := validator.ValidateCreate(ctx, obj)
			Expect(err).To(HaveOccurred())
		})

		It("should allow creation for valid JsonServer", func() {
			obj := &examplev1.JsonServer{
				ObjectMeta: metav1.ObjectMeta{
					Name: "app-valid",
				},
				Spec: examplev1.JsonServerSpec{
					JsonConfig: `{
						"people": [
							{ "id": 1, "name": "Alice" }
						]
					}`,
				},
			}

			_, err := validator.ValidateCreate(ctx, obj)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Context("ValidateUpdate", func() {
		It("should allow update even if jsonConfig becomes invalid", func() {
			oldObj := &examplev1.JsonServer{
				ObjectMeta: metav1.ObjectMeta{
					Name: "app-test",
				},
				Spec: examplev1.JsonServerSpec{
					JsonConfig: `{}`,
				},
			}

			newObj := &examplev1.JsonServer{
				ObjectMeta: metav1.ObjectMeta{
					Name: "app-test",
				},
				Spec: examplev1.JsonServerSpec{
					JsonConfig: `{ invalid json }`,
				},
			}

			_, err := validator.ValidateUpdate(ctx, oldObj, newObj)
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
