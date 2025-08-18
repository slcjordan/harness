package kube

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/kylelemons/godebug/diff"
	"github.com/slcjordan/harness"
	"github.com/slcjordan/harness/logger"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
)

type Differ struct {
	Clientset *kubernetes.Clientset
}

func (d *Differ) Handle(ctx context.Context, namespaces []string) ([]string, error) {
	var result []string
	var orig string

	for i, ns := range namespaces {
		curr, err := d.Clientset.AppsV1().Deployments(ns).List(ctx, metav1.ListOptions{})
		if err != nil {
			return nil, err
		}
		val, err := json.Marshal(curr)
		if err != nil {
			return nil, err
		}
		if i == 0 {
			orig = string(val)
		} else {
			unified := diff.Diff(orig, string(val))
			result = append(result, unified)
		}
	}
	return result, nil
}

type Lister struct {
	Clientset *kubernetes.Clientset
}

func shallowCopy[K comparable, V any](kv map[K]V) map[K]V {
	result := make(map[K]V)
	for k, v := range kv {
		result[k] = v
	}
	return result
}

func (l *Lister) Handle(ctx context.Context, namespaces []string) ([]harness.NamespacedObject, error) {
	var result []harness.NamespacedObject
	seen := make(map[string]bool)

	for i, ns := range namespaces {
		sts, err := l.Clientset.AppsV1().StatefulSets(ns).List(ctx, metav1.ListOptions{})
		if err != nil {
			return nil, err
		}
		extras := shallowCopy(seen)

		for _, curr := range sts.Items {
			name := curr.ObjectMeta.Name
			if i == 0 {
				seen[name] = true
			} else {
				if extras[name] {
					delete(extras, name)
				} else {
					result = append(result, harness.NamespacedObject{Namespace: "+" + ns, ID: name})
				}
			}
		}

		list, err := l.Clientset.AppsV1().Deployments(ns).List(ctx, metav1.ListOptions{})
		if err != nil {
			return nil, err
		}

		for _, curr := range list.Items {
			name := curr.ObjectMeta.Name
			if i == 0 {
				seen[name] = true
			} else {
				if extras[name] {
					delete(extras, name)
				} else {
					result = append(result, harness.NamespacedObject{Namespace: "+" + ns, ID: name})
				}
			}
		}

		for name := range extras {
			result = append(result, harness.NamespacedObject{Namespace: "-" + ns, ID: name})
		}
	}
	fmt.Println(result)
	return result, nil
}

type PortForwardDeploy[A, B any] struct {
	Config    *rest.Config
	Clientset *kubernetes.Clientset
	Handler   harness.Handler[A, B]
	Namespace string
	Name      string
	Ports     []string // "%s:%s" (local:remote) format
}

func (p *PortForwardDeploy[A, B]) Handle(ctx context.Context, input A) (B, error) {
	logger.Infof(ctx, "handling port-forward")
	var result B
	logger.Infof(ctx, "finding deployment")
	deploy, err := p.Clientset.AppsV1().Deployments(p.Namespace).Get(ctx, p.Name, metav1.GetOptions{})
	if err != nil {
		return result, fmt.Errorf("failed to get deployment: %w", err)
	}

	selector := metav1.FormatLabelSelector(&metav1.LabelSelector{MatchLabels: deploy.Spec.Selector.MatchLabels})

	logger.Infof(ctx, "listing pods")
	podList, err := p.Clientset.CoreV1().Pods(p.Namespace).List(ctx, metav1.ListOptions{
		LabelSelector: selector,
	})
	if err != nil {
		return result, fmt.Errorf("failed to list pods: %w", err)
	}

	if len(podList.Items) == 0 {
		return result, fmt.Errorf("no pods found for deployment %s", p.Name)
	}
	logger.Infof(ctx, "setting up round-tripper")
	transport, upgrader, err := spdy.RoundTripperFor(p.Config)
	if err != nil {
		return result, err
	}
	podName := podList.Items[0].Name
	logger.Infof(ctx, "portforward subresource")
	req := p.Clientset.CoreV1().
		RESTClient().
		Post().
		Resource("pods").
		Namespace(p.Namespace).
		Name(podName).
		SubResource("portforward")
	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, "POST", req.URL())

	stopChan := make(chan struct{}, 1)
	readyChan := make(chan struct{})
	defer close(stopChan)
	defer close(readyChan)
	logger.Infof(ctx, "new portforward")
	pf, err := portforward.New(dialer, p.Ports, stopChan, readyChan, os.Stdout, os.Stderr)
	if err != nil {
		return result, err
	}
	go func() {
		if err := pf.ForwardPorts(); err != nil {
			logger.Errorf(ctx, "while forwarding %s.%s[%s] (%v): %s", p.Namespace, p.Name, podName, p.Ports, err)
		}
	}()
	logger.Infof(ctx, "waiting for [%s]%s:%s to be ready", p.Namespace, podName, p.Ports)
	<-readyChan
	logger.Infof(ctx, "forwarding [%s]%s:%s", p.Namespace, podName, p.Ports)

	return p.Handler.Handle(ctx, input)
}
