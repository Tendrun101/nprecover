package main

import (
	"bytes"
	"context"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"os"
	"runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"strconv"

	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func main() {
	logf.SetLogger(zap.New())

	var log = logf.Log.WithName("builder-examples")

	mgr, err := manager.New(config.GetConfigOrDie(), manager.Options{})
	if err != nil {
		log.Error(err, "could not create manager")
		os.Exit(1)
	}

	deleteEventFilter := predicate.Funcs{
		DeleteFunc: func(deleteEvent event.DeleteEvent) bool {
			return deleteEvent.Object.GetDeletionTimestamp() != nil
		},
	}
	err = builder.
		ControllerManagedBy(mgr). // Create the ControllerManagedBy
		For(&corev1.Event{}).
		WithEventFilter(deleteEventFilter). // Filter Event
		Complete(&EventReconciler{})
	if err != nil {
		log.Error(err, "could not create controller")
		os.Exit(1)
	}

	if err := mgr.Start(signals.SetupSignalHandler()); err != nil {
		log.Error(err, "could not start manager")
		os.Exit(1)
	}
}

// EventReconciler is a simple ControllerManagedBy example implementation.
type EventReconciler struct {
	client.Client
}

// Reconcile Implement the business logic:
// This function will be called when there is a change to a ReplicaSet or a Pod with an OwnerReference
// to a ReplicaSet.
//
// * Read the ReplicaSet
// * Read the Pods
// * Set a Label on the ReplicaSet with the Pod count.
func (a *EventReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {

	fmt.Printf("the NamespaceName is %v\n", req.NamespacedName)
	event := &corev1.Event{}
	err := a.Get(ctx, req.NamespacedName, event)

	// 默认把Delete 事件的对象也会被放入队列，此时实际已Get 不到此实际对象
	if errors.IsNotFound(err) {
		fmt.Printf("%v not found, return\n", req.NamespacedName)
		return reconcile.Result{}, nil
	}
	if err != nil {
		return reconcile.Result{}, err
	}
	fmt.Printf("the new event is %v\n", event.Name)

	/*	fmt.Printf("list event ...")
		events := &corev1.EventList{}
		err = a.List(ctx, events, client.InNamespace(req.Namespace))
		if err != nil {
			return reconcile.Result{}, err
		}

		fmt.Printf("%d goroutine Reconcile ...\n", GetGID())
		for index,v := range events.Items{
			fmt.Printf("the %d event list is %v\n", index, v.Name)
		}
		fmt.Printf("\n\n\n\n\n")*/

	return reconcile.Result{}, nil
}

func (a *EventReconciler) InjectClient(c client.Client) error {
	a.Client = c
	return nil
}

func GetGID() uint64 {
	b := make([]byte, 64)
	b = b[:runtime.Stack(b, false)]
	b = bytes.TrimPrefix(b, []byte("goroutine "))
	b = b[:bytes.IndexByte(b, ' ')]
	n, _ := strconv.ParseUint(string(b), 10, 64)
	return n
}
