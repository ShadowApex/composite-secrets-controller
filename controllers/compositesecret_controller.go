/*
Copyright 2022.

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

package controllers

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	compositev1alpha1 "github.com/shadowapex/composite-secrets-controller/api/v1alpha1"
)

// CompositeSecretReconciler reconciles a CompositeSecret object
type CompositeSecretReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=composite.shadowblip.com,resources=compositesecrets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=composite.shadowblip.com,resources=compositesecrets/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=composite.shadowblip.com,resources=compositesecrets/finalizers,verbs=updatea
//+kubebuilder:rbac:groups=v1,resources=configmaps,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=v1,resources=secrets,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the CompositeSecret object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.2/pkg/reconcile
func (r *CompositeSecretReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("Running reconcile")
	defer func() { logger.Info("Reconcile completed") }()

	// Fetch the composite secret we're reconciling
	compositeSecret := &compositev1alpha1.CompositeSecret{}
	err := r.Get(ctx, req.NamespacedName, compositeSecret)
	if err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// Lookup and map all replacements
	replacements := map[string]string{}
	for varName, replace := range compositeSecret.Spec.Replacements {
		if replace.ConfigMapRef != nil && replace.SecretRef != nil {
			return ctrl.Result{}, fmt.Errorf("replacement cannot specify both configmap and secret")
		}
		if replace.ConfigMapRef != nil {
			data, err := r.getConfigMapKey(ctx, replace.ConfigMapRef)
			if err != nil {
				return ctrl.Result{}, err
			}
			replacements[varName] = data
		}
		if replace.SecretRef != nil {
			data, err := r.getSecretKey(ctx, replace.SecretRef)
			if err != nil {
				return ctrl.Result{}, err
			}
			replacements[varName] = data
		}
	}

	// Render the replaements
	data, err := r.renderReplacements(compositeSecret, replacements)
	if err != nil {
		return ctrl.Result{}, err
	}

	// Check if the secret already exists, if not create a new secret.
	found := &v1.Secret{}
	err = r.Get(ctx, types.NamespacedName{Name: compositeSecret.Name, Namespace: compositeSecret.Namespace}, found)
	if err != nil {
		if errors.IsNotFound(err) {
			// Define and create a new secret.
			secret := r.secretForCompositeSecret(compositeSecret, data)
			if err = r.Create(ctx, secret); err != nil {
				return ctrl.Result{}, err
			}
			return ctrl.Result{Requeue: true}, nil
		} else {
			return ctrl.Result{}, err
		}
	}

	// Ensure the secret data is the same as the rendered spec.
	if !reflect.DeepEqual(data, found.Data) {
		found.Data = data
		if err = r.Update(ctx, found); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{Requeue: true}, nil
	}

	// Ensure our labels are synced
	if !reflect.DeepEqual(compositeSecret.ObjectMeta.Labels, found.ObjectMeta.Labels) {
		found.ObjectMeta.Labels = r.applyLabels(compositeSecret.ObjectMeta.Labels)
		if err = r.Update(ctx, found); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{Requeue: true}, nil
	}

	return ctrl.Result{}, nil
}

// getConfigMapKey will return the data in the given configmap reference.
func (r *CompositeSecretReconciler) getConfigMapKey(ctx context.Context, replace *compositev1alpha1.ObjectRef) (string, error) {
	key := replace.Key
	configMap := &v1.ConfigMap{}
	name := types.NamespacedName{
		Name:      replace.Name,
		Namespace: replace.Namespace,
	}
	err := r.Get(ctx, name, configMap)
	if err != nil {
		return "", err
	}
	data, ok := configMap.Data[key]
	if !ok {
		return "", fmt.Errorf("no key '%v' found in %v", key, name)
	}
	return data, nil
}

// getSecretKey will return the data in the given secret reference.
func (r *CompositeSecretReconciler) getSecretKey(ctx context.Context, replace *compositev1alpha1.ObjectRef) (string, error) {
	key := replace.Key
	secret := &v1.Secret{}
	name := types.NamespacedName{
		Name:      replace.Name,
		Namespace: replace.Namespace,
	}
	err := r.Get(ctx, name, secret)
	if err != nil {
		return "", fmt.Errorf("unable to get secret: %v", err)
	}
	data, ok := secret.Data[key]
	if !ok {
		return "", fmt.Errorf("no key '%v' found in %v", key, name)
	}

	return string(data), nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *CompositeSecretReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&compositev1alpha1.CompositeSecret{}).
		Owns(&v1.Secret{}).
		Complete(r)
}

// renderReplacements will return the data of the given CompositeSecret template
// with applied replacements.
func (r *CompositeSecretReconciler) renderReplacements(m *compositev1alpha1.CompositeSecret, replacements map[string]string) (map[string][]byte, error) {
	strData := map[string]string{}
	data := map[string][]byte{}

	// Decode Data from base64
	if m.Spec.Template.StringData != nil {
		strData = m.Spec.Template.StringData
	}

	// Apply replacements to all keys in the secret
	for key, value := range strData {
		for varName, replace := range replacements {
			value = strings.ReplaceAll(value, varName, replace)
		}
		data[key] = []byte(value)
	}

	return data, nil
}

// secretForCompositeSecret returns a Secret object for data from m.
func (r *CompositeSecretReconciler) secretForCompositeSecret(m *compositev1alpha1.CompositeSecret, data map[string][]byte) *v1.Secret {
	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.Name,
			Namespace: m.Namespace,
			Labels:    r.applyLabels(m.Labels),
		},
		Immutable: m.Spec.Template.Immutable,
		Type:      m.Spec.Template.Type,
		Data:      data,
	}

	// Set CompositeSecret instance as the owner and controller.
	// NOTE: calling SetControllerReference, and setting owner references in
	// general, is important as it allows deleted objects to be garbage collected.
	controllerutil.SetControllerReference(m, secret, r.Scheme)
	return secret
}

// applyLabels will return the given labels with controller labels added
func (r *CompositeSecretReconciler) applyLabels(lbls map[string]string) map[string]string {
	labels := map[string]string{}
	labels["composite.shadowblip.com/managed-by"] = "composite-secrets-controller"
	for key, value := range lbls {
		labels[key] = value
	}
	return labels
}
