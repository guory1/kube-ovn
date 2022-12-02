package controller

import (
	"github.com/kubeovn/kube-ovn/pkg/util"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/klog/v2"
	"time"
)

func (c *Controller) runProtectLoadBalancer() {
	klog.Info("Starting protect load_balancer")
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
loop:
	for {
		select {
		case <-ticker.C:
			err := c.protectLoadBalancer()
			if err != nil {
				klog.Errorf("failed to protect load_balancer: %v", err)
				break loop
			}
		}
	}
	return
}

func (c *Controller) protectLoadBalancer() error {
	klog.Infof("start to protect loadbalancers")
	if c.config.EnableLb {
		vpcs, err := c.vpcsLister.List(labels.Everything())
		if err != nil {
			return err
		}
		for _, orivpc := range vpcs {
			vpc := orivpc.DeepCopy()
			if value, ok := vpc.Annotations[util.VpcEnableLbAnnotation]; !ok || value != util.VpcAnnotationEnableOn {
				continue
			}
			// check load_balancer
			for _, subnetName := range vpc.Status.Subnets {
				err = c.ovnClient.CheckAndAddLbToLogicalSwitch(
					vpc.Status.TcpLoadBalancer,
					vpc.Status.TcpSessionLoadBalancer,
					vpc.Status.UdpLoadBalancer,
					vpc.Status.UdpSessionLoadBalancer,
					subnetName)
				if err != nil {
					klog.Errorf("failed to check lb of ls %s, %v", subnetName, err)
					return err
				}
			}
		}
	}

	return nil
}
