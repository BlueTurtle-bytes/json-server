/*
Copyright 2026.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	examplev1 "github.com/BlueTurtle-bytes/json-server/api/v1"
)

// nolint:unused
// log is for logging in this package.
var jsonserverlog = logf.Log.WithName("jsonserver-resource")

// SetupJsonServerWebhookWithManager registers the webhook for JsonServer in the manager.
func SetupJsonServerWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr, &examplev1.JsonServer{}).
		WithValidator(&JsonServerCustomValidator{}).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
// NOTE: If you want to customise the 'path', use the flags '--defaulting-path' or '--validation-path'.
// +kubebuilder:webhook:path=/validate-example-com-v1-jsonserver,mutating=false,failurePolicy=fail,sideEffects=None,groups=example.com,resources=jsonservers,verbs=create;update,versions=v1,name=vjsonserver-v1.kb.io,admissionReviewVersions=v1

// JsonServerCustomValidator struct is responsible for validating the JsonServer resource
// when it is created, updated, or deleted.
//
// NOTE: The +kubebuilder:object:generate=false marker prevents controller-gen from generating DeepCopy methods,
// as this struct is used only for temporary operations and does not need to be deeply copied.
type JsonServerCustomValidator struct {
	// TODO(user): Add more fields as needed for validation
}

// ValidateCreate implements webhook.CustomValidator so a webhook will be registered for the type JsonServer.
func (v *JsonServerCustomValidator) ValidateCreate(_ context.Context, obj *examplev1.JsonServer) (admission.Warnings, error) {

	if !strings.HasPrefix(obj.Name, "app-") {
		return nil, fmt.Errorf("Error: metadata.name must start with app-")
	}

	var js any
	if err := json.Unmarshal([]byte(obj.Spec.JsonConfig), &js); err != nil {
		return nil, fmt.Errorf("Error: spec.jsonConfig is not a valid json object")
	}

	return nil, nil
}

// ValidateUpdate implements webhook.CustomValidator so a webhook will be registered for the type JsonServer.
func (v *JsonServerCustomValidator) ValidateUpdate(_ context.Context, oldObj, newObj *examplev1.JsonServer) (admission.Warnings, error) {

	// dont allow allow name changes
	if oldObj.Name != newObj.Name {
		return nil, fmt.Errorf("metadata.name is immutable")
	}

	// Allow invalid JSON updates
	// Controller will detect and update status

	return nil, nil
}

// ValidateDelete implements webhook.CustomValidator so a webhook will be registered for the type JsonServer.
func (v *JsonServerCustomValidator) ValidateDelete(_ context.Context, obj *examplev1.JsonServer) (admission.Warnings, error) {
	jsonserverlog.Info("Validation for JsonServer upon deletion", "name", obj.GetName())

	// TODO(user): fill in your validation logic upon object deletion.

	return nil, nil
}
