/*
Copyright 2021.

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
	"time"

	"github.com/blang/semver"
	"github.com/giantswarm/k8smetadata/pkg/label"
	"github.com/giantswarm/microerror"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	capi "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/cluster-api/util/patch"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/giantswarm/encryption-provider-operator/pkg/encryption"
	"github.com/giantswarm/encryption-provider-operator/pkg/key"
)

// ClusterReconciler reconciles a Cluster object
type ClusterReconciler struct {
	AppCatalog               string
	DefaultKeyRotationPeriod time.Duration
	RegistryDomain           string
	FromReleaseVersion       string

	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=cluster.x-k8s.io,resources=cluster,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cluster.x-k8s.io,resources=cluster/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=cluster.x-k8s.io,resources=cluster/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
func (r *ClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var err error
	logger := r.Log.WithValues("namespace", req.Namespace, "cluster", req.Name)

	cluster := &capi.Cluster{}
	if err := r.Get(ctx, req.NamespacedName, cluster); err != nil {
		logger.Error(err, "Cluster does not exist")
		return ctrl.Result{}, microerror.Mask(err)
	}

	// if the cluster CR has a old GS release label we check if the release version is old enought for encryption operator,
	// otherwise ignore the CR
	if v, ok := cluster.Labels[label.ReleaseVersion]; ok {
		version, err := semver.Parse(v)
		if err != nil {
			return ctrl.Result{}, microerror.Mask(err)
		}

		fromRelease, err := semver.Parse(r.FromReleaseVersion)
		if err != nil {
			return ctrl.Result{}, microerror.Mask(err)
		}

		if version.LT(fromRelease) {
			// release is older than supported version, ignore the CR
			logger.Info(fmt.Sprintf("cluster is running old release %s which does not support encryption-provider-operaotr, ignoring the CR", v))
			return ctrl.Result{}, nil
		}
	} else {
		logger.Info("did not found release label on cluster CR, assuming CAPI release")
	}

	var encryptionService *encryption.Service
	{
		c := encryption.Config{
			AppCatalog:               r.AppCatalog,
			Cluster:                  cluster,
			CtrlClient:               r.Client,
			DefaultKeyRotationPeriod: r.DefaultKeyRotationPeriod,
			RegistryDomain:           r.RegistryDomain,
			Logger:                   logger,
		}

		encryptionService, err = encryption.New(c)
		if err != nil {
			logger.Error(err, "failed to create encryption service")
			return ctrl.Result{}, microerror.Mask(err)
		}
	}
	patchHelper, err := patch.NewHelper(cluster, r.Client)
	if err != nil {
		return ctrl.Result{}, err
	}

	if cluster.DeletionTimestamp != nil {
		// clean
		err = encryptionService.Delete()
		if err != nil {
			logger.Error(err, "failed to clean resources")
			return ctrl.Result{}, microerror.Mask(err)
		}
		// remove finalizer from Cluster
		controllerutil.RemoveFinalizer(cluster, key.FinalizerName)
		err = patchHelper.Patch(ctx, cluster)
		if err != nil {
			logger.Error(err, "failed to remove finalizer on Cluster CR")
			return ctrl.Result{}, microerror.Mask(err)
		}
		// resource was cleaned up, we dont need to reconcile again
		return ctrl.Result{}, nil

	} else {
		// add finalizer to AWSMachineTemplate
		controllerutil.AddFinalizer(cluster, key.FinalizerName)
		patchHelper.Patch(ctx, cluster)
		if err != nil {
			logger.Error(err, "failed to add finalizer on Cluster CR")
			return ctrl.Result{}, microerror.Mask(err)
		}

		// reconcile
		err = encryptionService.Reconcile()
		if err != nil {
			logger.Error(err, "failed to reconcile resource")
			return ctrl.Result{}, microerror.Mask(err)
		}
	}

	return ctrl.Result{
		Requeue:      true,
		RequeueAfter: time.Minute * 5,
	}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&capi.Cluster{}).
		Complete(r)
}
