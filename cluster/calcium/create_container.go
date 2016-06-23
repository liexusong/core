package calcium

import (
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	enginetypes "github.com/docker/engine-api/types"
	enginecontainer "github.com/docker/engine-api/types/container"
	enginenetwork "github.com/docker/engine-api/types/network"
	engineslice "github.com/docker/engine-api/types/strslice"
	"github.com/docker/go-units"
	"gitlab.ricebook.net/platform/core/types"
	"gitlab.ricebook.net/platform/core/utils"
	"golang.org/x/net/context"
)

// Create Container
// Use specs and options to create
// TODO what about networks?
func (c *Calcium) CreateContainer(specs types.Specs, opts *types.DeployOptions) (chan *types.CreateContainerMessage, error) {
	ch := make(chan *types.CreateContainerMessage)

	result, err := c.prepareNodes(opts.Podname, opts.CPUQuota, opts.Count)
	if err != nil {
		return ch, err
	}
	if len(result) == 0 {
		return ch, fmt.Errorf("Not enough resource to create container")
	}

	// check total count in case scheduler error
	totalCount := 0
	for _, cores := range result {
		totalCount = totalCount + len(cores)
	}
	if totalCount != opts.Count {
		return ch, fmt.Errorf("Count mismatch (opt.Count %q, total %q), maybe scheduler error?", opts.Count, totalCount)
	}

	go func() {
		wg := sync.WaitGroup{}
		wg.Add(len(result))

		// do deployment
		for nodename, cpumap := range result {
			go func(nodename string, cpumap []types.CPUMap, opts *types.DeployOptions) {
				defer wg.Done()

				for _, m := range c.doCreateContainer(nodename, cpumap, specs, opts) {
					ch <- m
				}
			}(nodename, cpumap, opts)
		}

		wg.Wait()
		close(ch)
	}()

	return ch, nil
}

func makeCPUMap(nodes []*types.Node) map[string]types.CPUMap {
	r := make(map[string]types.CPUMap)
	for _, node := range nodes {
		r[node.Name] = node.CPU
	}
	return r
}

// Prepare nodes for deployment.
// Later if any error occurs, these nodes can be restored.
func (c *Calcium) prepareNodes(podname string, quota float64, num int) (map[string][]types.CPUMap, error) {
	// TODO use distributed lock on podname instead
	c.Lock()
	defer c.Unlock()

	q := int(quota)
	r := make(map[string][]types.CPUMap)

	nodes, err := c.ListPodNodes(podname)
	if err != nil {
		return r, err
	}

	// if public, use only public nodes
	if q == 0 {
		nodes = filterNodes(nodes, true)
	} else {
		nodes = filterNodes(nodes, false)
	}

	cpumap := makeCPUMap(nodes)
	r, err = c.scheduler.SelectNodes(cpumap, q, num)
	if err != nil {
		return r, err
	}

	// if quota is set to 0
	// then no cpu is required
	if q > 0 {
		// cpus remained
		// update data to etcd
		// `SelectNodes` reduces count in cpumap
		for _, node := range nodes {
			node.CPU = cpumap[node.Name]
			// ignore error
			c.store.UpdateNode(node)
		}
	}

	return r, err
}

// filter nodes
// public is the flag
func filterNodes(nodes []*types.Node, public bool) []*types.Node {
	rs := []*types.Node{}
	for _, node := range nodes {
		if node.Public == public {
			rs = append(rs, node)
		}
	}
	return rs
}

// Pull an image
// Blocks until it finishes.
func pullImage(node *types.Node, image string) error {
	if image == "" {
		return fmt.Errorf("No image found for version")
	}

	resp, err := node.Engine.ImagePull(context.Background(), image, enginetypes.ImagePullOptions{})
	if err != nil {
		return err
	}
	ensureReaderClosed(resp)
	return nil
}

func (c *Calcium) doCreateContainer(nodename string, cpumap []types.CPUMap, specs types.Specs, opts *types.DeployOptions) []*types.CreateContainerMessage {
	ms := make([]*types.CreateContainerMessage, len(cpumap))
	for i := 0; i < len(ms); i++ {
		ms[i] = &types.CreateContainerMessage{}
	}

	node, err := c.GetNode(opts.Podname, nodename)
	if err != nil {
		return ms
	}

	if err := pullImage(node, opts.Image); err != nil {
		return ms
	}

	for i, quota := range cpumap {
		config, hostConfig, networkConfig, containerName, err := c.makeContainerOptions(quota, specs, opts)
		if err != nil {
			log.Errorf("error when creating CreateContainerOptions, %v", err)
			c.releaseQuota(node, quota)
			continue
		}

		container, err := node.Engine.ContainerCreate(context.Background(), config, hostConfig, networkConfig, containerName)
		if err != nil {
			log.Errorf("error when creating container, %v", err)
			c.releaseQuota(node, quota)
			continue
		}

		err = node.Engine.ContainerStart(context.Background(), container.ID, enginetypes.ContainerStartOptions{})
		if err != nil {
			log.Errorf("error when starting container, %v", err)
			c.releaseQuota(node, quota)
			go node.Engine.ContainerRemove(context.Background(), container.ID, enginetypes.ContainerRemoveOptions{})
			continue
		}

		info, err := node.Engine.ContainerInspect(context.Background(), container.ID)
		if err != nil {
			log.Errorf("error when inspecting container, %v", err)
			c.releaseQuota(node, quota)
			continue
		}

		_, err = c.store.AddContainer(info.ID, opts.Podname, node.Name, containerName, quota)
		if err != nil {
			c.releaseQuota(node, quota)
			continue
		}

		ms[i] = &types.CreateContainerMessage{
			Podname:       opts.Podname,
			Nodename:      node.Name,
			ContainerID:   info.ID,
			ContainerName: containerName,
			Success:       true,
			CPU:           quota,
		}
	}

	return ms
}

// When deploy on a public host
// quota is set to 0
// no need to update this to etcd (save 1 time write on etcd)
func (c *Calcium) releaseQuota(node *types.Node, quota types.CPUMap) {
	if quota.Total() == 0 {
		return
	}
	node.CPU.Add(quota)
	c.store.UpdateNode(node)
}

func (c *Calcium) makeContainerOptions(quota map[string]int, specs types.Specs, opts *types.DeployOptions) (
	*enginecontainer.Config,
	*enginecontainer.HostConfig,
	*enginenetwork.NetworkingConfig,
	string,
	error) {

	entry, ok := specs.Entrypoints[opts.Entrypoint]
	if !ok {
		return nil, nil, nil, "", fmt.Errorf("Entrypoint %q not found in image %q", opts.Entrypoint, opts.Image)
	}

	// command
	slices := strings.Split(entry.Command, " ")
	starter, needNetwork := "launcher", "network"
	if !opts.Raw {
		if entry.Privileged != "" {
			starter = "launcheroot"
		}
		if len(opts.Networks) == 0 {
			needNetwork = "nonetwork"
		}
		slices = append([]string{fmt.Sprintf("/usr/local/bin/%s", starter), needNetwork}, slices...)
	}
	cmd := engineslice.StrSlice(slices)

	// calculate CPUShares and CPUSet
	// scheduler won't return more than 1 share quota
	// so the smallest share is the share numerator
	shareQuota := 10
	labels := []string{}
	for label, share := range quota {
		labels = append(labels, label)
		if share < shareQuota {
			shareQuota = share
		}
	}
	cpuShares := int64(float64(shareQuota) / float64(10) * float64(1024))
	cpuSetCpus := strings.Join(labels, ",")

	// env
	env := append(opts.Env, fmt.Sprintf("APP_NAME=%s", specs.Appname))
	env = append(env, fmt.Sprintf("ERU_POD=%s", opts.Podname))

	// volumes and binds
	volumes := make(map[string]struct{})
	volumes["/writable-proc/sys"] = struct{}{}

	binds := []string{}
	binds = append(binds, "/proc/sys:/writable-proc/sys:ro")

	// add permdir to container
	if entry.PermDir {
		permDir := filepath.Join("/", specs.Appname, "permdir")
		permDirHost := filepath.Join(c.config.PermDir, specs.Appname)
		volumes[permDir] = struct{}{}

		binds = append(binds, strings.Join([]string{permDirHost, permDir, "rw"}, ":"))
		env = append(env, fmt.Sprintf("ERU_PERMDIR=%s", permDir))
	}

	for _, volume := range specs.Volumes {
		volumes[volume] = struct{}{}
	}

	var mode string
	for hostPath, bind := range specs.Binds {
		if bind.ReadOnly {
			mode = "ro"
		} else {
			mode = "rw"
		}
		binds = append(binds, strings.Join([]string{hostPath, bind.InContainerPath, mode}, ":"))
	}

	// log config
	logConfig := c.config.Docker.LogDriver
	if entry.LogConfig == "json-file" {
		logConfig = "json-file"
	}

	// working dir is /:appname if it's not deployed as raw app
	workingDir := "/"
	if !opts.Raw {
		workingDir = "/" + specs.Appname
	}

	// CapAdd and Privileged
	capAdd := []string{}
	if entry.Privileged == "__super__" {
		capAdd = append(capAdd, "SYS_ADMIN")
	}

	// ulimit
	ulimits := []*units.Ulimit{&units.Ulimit{Name: "nofile", Soft: 65535, Hard: 65535}}

	// name
	suffix := utils.RandomString(6)
	containerName := strings.Join([]string{specs.Appname, opts.Entrypoint, suffix}, "_")

	config := &enginecontainer.Config{
		Env:             env,
		Cmd:             cmd,
		Image:           opts.Image,
		Volumes:         volumes,
		WorkingDir:      workingDir,
		NetworkDisabled: false,
		Labels:          make(map[string]string),
	}
	hostConfig := &enginecontainer.HostConfig{
		Binds:         binds,
		LogConfig:     enginecontainer.LogConfig{Type: logConfig},
		NetworkMode:   enginecontainer.NetworkMode(entry.NetworkMode),
		RestartPolicy: enginecontainer.RestartPolicy{Name: entry.RestartPolicy, MaximumRetryCount: 3},
		CapAdd:        engineslice.StrSlice(capAdd),
		ExtraHosts:    entry.ExtraHosts,
		Privileged:    entry.Privileged != "",
		Resources: enginecontainer.Resources{
			CPUShares:  cpuShares,
			CpusetCpus: cpuSetCpus,
			Ulimits:    ulimits,
		},
	}
	// this is empty because we don't use any plugin for Docker
	networkConfig := &enginenetwork.NetworkingConfig{}
	return config, hostConfig, networkConfig, containerName, nil
}

// Upgrade containers
// Use image to run these containers, and copy their settings
// Note, if the image is not correct, container will be started incorrectly
// TODO what about networks?
func (c *Calcium) UpgradeContainer(ids []string, image string) (chan *types.UpgradeContainerMessage, error) {
	ch := make(chan *types.UpgradeContainerMessage)

	if len(ids) == 0 {
		return ch, fmt.Errorf("No container ids given")
	}

	containers, err := c.GetContainers(ids)
	if err != nil {
		return ch, err
	}

	containerMap := make(map[string][]*types.Container)
	for _, container := range containers {
		containerMap[container.Nodename] = append(containerMap[container.Nodename], container)
	}

	go func() {
		wg := sync.WaitGroup{}
		wg.Add(len(containerMap))

		for _, containers := range containerMap {
			go func(containers []*types.Container, image string) {
				defer wg.Done()

				for _, m := range c.doUpgradeContainer(containers, image) {
					ch <- m
				}
			}(containers, image)

		}

		wg.Wait()
		close(ch)
	}()

	return ch, nil
}

// upgrade containers on the same node
func (c *Calcium) doUpgradeContainer(containers []*types.Container, image string) []*types.UpgradeContainerMessage {
	ms := make([]*types.UpgradeContainerMessage, len(containers))
	for i := 0; i < len(ms); i++ {
		ms[i] = &types.UpgradeContainerMessage{}
	}

	// TODO ugly
	// use the first container to get node
	// since all containers here must locate on the same node and pod
	t := containers[0]
	node, err := c.GetNode(t.Podname, t.Nodename)
	if err != nil {
		return ms
	}

	// prepare new image
	if err := pullImage(node, image); err != nil {
		return ms
	}

	imagesToDelete := make(map[string]struct{})
	engine := node.Engine

	for i, container := range containers {
		info, err := container.Inspect()
		if err != nil {
			ms[i].Error = err.Error()
			continue
		}

		// stops the old container
		timeout := 5 * time.Second
		err = engine.ContainerStop(context.Background(), info.ID, &timeout)
		if err != nil {
			ms[i].Error = err.Error()
			continue
		}

		// copy config from old container
		// and of course with a new name
		config, hostConfig, networkConfig, containerName, err := makeContainerConfig(info, image)
		if err != nil {
			ms[i].Error = err.Error()
			continue
		}

		// create a container with old config and a new name
		newContainer, err := engine.ContainerCreate(context.Background(), config, hostConfig, networkConfig, containerName)
		if err != nil {
			ms[i].Error = err.Error()
			continue
		}

		// start this new container
		err = engine.ContainerStart(context.Background(), newContainer.ID, enginetypes.ContainerStartOptions{})
		if err != nil {
			go engine.ContainerRemove(context.Background(), newContainer.ID, enginetypes.ContainerRemoveOptions{})
			ms[i].Error = err.Error()
			continue
		}

		// test if container is correctly started
		// if not, restore the old container
		newInfo, err := engine.ContainerInspect(context.Background(), newContainer.ID)
		if err != nil {
			ms[i].Error = err.Error()
			// restart the old container
			engine.ContainerStart(context.Background(), info.ID, enginetypes.ContainerStartOptions{})
			continue
		}

		// if so, add a new container in etcd
		_, err = c.store.AddContainer(newInfo.ID, container.Podname, container.Nodename, containerName, container.CPU)
		if err != nil {
			ms[i].Error = err.Error()
			continue
		}

		// remove the old container on node
		rmOpts := enginetypes.ContainerRemoveOptions{
			RemoveVolumes: true,
			Force:         true,
		}
		err = engine.ContainerRemove(context.Background(), info.ID, rmOpts)
		if err != nil {
			ms[i].Error = err.Error()
			continue
		}

		imagesToDelete[info.Image] = struct{}{}

		// remove the old container in etcd
		err = c.store.RemoveContainer(info.ID)
		if err != nil {
			ms[i].Error = err.Error()
			continue
		}

		// send back the message
		ms[i].ContainerID = info.ID
		ms[i].NewContainerID = newContainer.ID
		ms[i].NewContainerName = containerName
		ms[i].Success = true
	}

	// clean all the container images
	go func() {
		rmiOpts := enginetypes.ImageRemoveOptions{
			Force:         false,
			PruneChildren: true,
		}
		for image, _ := range imagesToDelete {
			engine.ImageRemove(context.Background(), image, rmiOpts)
		}
	}()
	return ms
}
