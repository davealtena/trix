package trivy

import (
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"

	"github.com/davealtena/trix/internal/k8s"
)

// Client wraps k8s.Client for Trivy-specific operations
type Client struct {
	k8sClient     *k8s.Client
	dynamicClient dynamic.Interface
	clientset     *kubernetes.Clientset
}

// NewClient creates a Trivy client from a k8s client
func NewClient(k8sClient *k8s.Client) *Client {
	return &Client{
		k8sClient:     k8sClient,
		dynamicClient: k8sClient.DynamicClient(),
		clientset:     k8sClient.Clientset(),
	}
}

// K8sClient returns the underlying k8s client
func (c *Client) K8sClient() *k8s.Client {
	return c.k8sClient
}
