package calcium

import (
	"context"

	"github.com/projecteru2/core/cluster"
	"github.com/projecteru2/core/types"
	"github.com/sanity-io/litter"
	log "github.com/sirupsen/logrus"
)

// AddNode add a node in pod
func (c *Calcium) AddNode(ctx context.Context, nodename, endpoint, podname, ca, cert, key string,
	cpu, share int, memory, storage int64, labels map[string]string,
	numa types.NUMA, numaMemory types.NUMAMemory) (*types.Node, error) {
	return c.store.AddNode(ctx, nodename, endpoint, podname, ca, cert, key, cpu, share, memory, storage, labels, numa, numaMemory)
}

// RemoveNode remove a node
func (c *Calcium) RemoveNode(ctx context.Context, podname, nodename string) error {
	return c.withNodeLocked(ctx, podname, nodename, func(node *types.Node) error {
		return c.store.RemoveNode(ctx, node)
	})
}

// ListPodNodes list nodes belong to pod
func (c *Calcium) ListPodNodes(ctx context.Context, podname string, labels map[string]string, all bool) ([]*types.Node, error) {
	return c.store.GetNodesByPod(ctx, podname, labels, all)
}

// GetNode get node
func (c *Calcium) GetNode(ctx context.Context, nodename string) (*types.Node, error) {
	return c.store.GetNode(ctx, nodename)
}

// GetNodes get nodes
func (c *Calcium) GetNodes(ctx context.Context, podname, nodename string, labels map[string]string, all bool) ([]*types.Node, error) {
	var ns []*types.Node
	var err error
	if nodename == "" {
		ns, err = c.ListPodNodes(ctx, podname, labels, all)
	} else {
		var node *types.Node
		node, err = c.GetNode(ctx, nodename)
		ns = []*types.Node{node}
	}
	return ns, err
}

// SetNode set node available or not
func (c *Calcium) SetNode(ctx context.Context, opts *types.SetNodeOptions) (*types.Node, error) {
	var n *types.Node
	return n, c.withNodeLocked(ctx, opts.Podname, opts.Nodename, func(node *types.Node) error {
		n = node
		litter.Dump(opts)
		// status
		switch opts.Status {
		case cluster.NodeUp:
			n.Available = true
		case cluster.NodeDown:
			n.Available = false
			containers, err := c.store.ListNodeContainers(ctx, opts.Nodename, nil)
			if err != nil {
				return err
			}
			for _, container := range containers {
				if container.StatusMeta == nil {
					container.StatusMeta = &types.StatusMeta{ID: container.ID}
				}
				container.StatusMeta.Running = false
				container.StatusMeta.Healthy = false

				// mark container which belongs to this node as unhealthy
				if err = c.store.SetContainerStatus(ctx, container, 0); err != nil {
					log.Errorf("[SetNodeAvailable] Set container %s on node %s inactive failed %v", container.ID, opts.Nodename, err)
				}
			}
		}
		// update key value
		if len(opts.Labels) != 0 {
			n.Labels = opts.Labels
		}
		// update numa
		if len(opts.NUMA) != 0 {
			n.NUMA = types.NUMA(opts.NUMA)
		}
		// update numa memory
		for numaNode, memoryDelta := range opts.DeltaNUMAMemory {
			if _, ok := n.NUMAMemory[numaNode]; ok {
				n.NUMAMemory[numaNode] += memoryDelta
				n.InitNUMAMemory[numaNode] += memoryDelta
				if n.NUMAMemory[numaNode] < 0 {
					return types.ErrBadMemory
				}
			}
		}
		if opts.DeltaStorage != 0 {
			// update storage
			n.StorageCap += opts.DeltaStorage
			n.InitStorageCap += opts.DeltaStorage
			if n.StorageCap < 0 {
				return types.ErrBadStorage
			}
		}
		if opts.DeltaMemory != 0 {
			// update memory
			n.MemCap += opts.DeltaMemory
			n.InitMemCap += opts.DeltaMemory
			if n.MemCap < 0 {
				return types.ErrBadStorage
			}
		}
		// update cpu
		for cpuID, cpuShare := range opts.DeltaCPU {
			if _, ok := n.CPU[cpuID]; !ok && cpuShare > 0 { // 增加了 CPU
				n.CPU[cpuID] = cpuShare
				n.InitCPU[cpuID] = cpuShare
			} else if ok && cpuShare == 0 { // 删掉 CPU
				delete(n.CPU, cpuID)
				delete(n.InitCPU, cpuID)
			} else if ok { // 减少份数
				n.CPU[cpuID] += cpuShare
				n.InitCPU[cpuID] += cpuShare
				if n.CPU[cpuID] < 0 {
					return types.ErrBadCPU
				}
			}
		}
		return c.store.UpdateNode(ctx, n)
	})
}