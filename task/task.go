package task

import (
	"errors"
	"flag"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/gawkermedia/ecs/cli"
)

var cliClusterName string
var cliContainerInstance string
var cliDesiredStatus string
var cliStatus string
var cliFamily string
var cliMaxResults int64
var cliMaxTries int
var cliTimeout int64
var cliServiceName string
var cliTaskDef string
var cliTasks string
var cliStartedby string
var cliContainerPort int64
var cliHostPort int64
var cliMountPath string
var cliSourceVolume string
var cliMountReadOnly bool
var cliImage string
var cliCPU int64
var cliMemory int64
var cliEssential bool
var cliLinks string

func containerDef(family *string, containerPort *int64, hostPort *int64, image *string, cpu *int64, memory *int64, essential bool, links []*string) *ecs.ContainerDefinition {
	portMapping := &ecs.PortMapping{
		ContainerPort: containerPort,
		HostPort:      hostPort,
		Protocol:      aws.String("tcp"),
	}
	mountPoint := &ecs.MountPoint{
		ContainerPath: aws.String("/var/www/" + (*family) + "-vol"),
		SourceVolume:  aws.String(*family + "-vol"),
		ReadOnly:      aws.Bool(false),
	}
	c := ecs.ContainerDefinition{}

	c.Name = aws.String(*family + "-app")
	c.Image = image
	c.Cpu = cpu
	c.Memory = memory
	c.Essential = aws.Bool(essential)
	c.Links = links
	if *hostPort != 0 {
		c.PortMappings = []*ecs.PortMapping{portMapping}
	}
	c.MountPoints = []*ecs.MountPoint{mountPoint}
	return &c
}

// Definition Task definition
func Definition(family *string, containerPort *int64, hostPort *int64, image *string, cpu *int64, memory *int64, essential bool, links []*string) *ecs.RegisterTaskDefinitionInput {
	return &ecs.RegisterTaskDefinitionInput{
		ContainerDefinitions: []*ecs.ContainerDefinition{
			containerDef(family, containerPort, hostPort, image, cpu, memory, essential, links),
		},
		Family: family,
		Volumes: []*ecs.Volume{
			{
				Name: aws.String(*family + "-vol"),
			},
		},
	}
}

// RegisterTask Register a new version of the task definition
func RegisterTask(svc *ecs.ECS, params *ecs.RegisterTaskDefinitionInput) (*ecs.RegisterTaskDefinitionOutput, error) {
	return svc.RegisterTaskDefinition(params)
}

func cliRegisterTaskParams(args []string) *flag.FlagSet {
	var c = cli.Get("", args)
	c.Int64Var(&cliContainerPort, "container-port", 9000, "The port number on the container that is bound to the user-specified or automatically assigned host port. If you specify a container port and not a host port, your container will automatically receive a host port in the ephemeral port range (for more information, see hostPort)")
	c.Int64Var(&cliHostPort, "host-port", 80, "The port number on the container instance to reserve for your container. You can specify a non-reserved host port for your container port mapping, or you can omit the hostPort (or set it to 0) while specifying a containerPort and your container will automatically receive a port in the ephemeral port range for your container instance operating system and Docker version.")
	c.StringVar(&cliFamily, "family", "", "The name of the family with which to filter the list-tasks results. Specifying a family limits the results to tasks that belong to that family.")
	c.StringVar(&cliMountPath, "mount-path", "", "The path on the container to mount the host volume at.")
	c.StringVar(&cliSourceVolume, "source-volume", "", "The name of the volume to mount.")
	c.BoolVar(&cliMountReadOnly, "mount-read-only", false, "If this value is true, the container has read-only access to the volume. If this value is false, then the container can write to the volume. The default value is false.")
	c.StringVar(&cliImage, "image", "", "The image used to start a container. This string is passed directly to the Docker daemon. Images in the Docker Hub registry are available by default. Other repositories are specified with repository-url/image:tag.")
	c.Int64Var(&cliCPU, "cpu", 1024, "The number of cpu units reserved for the container. A container instance has 1,024 cpu units for every CPU core. This parameter specifies the minimum amount of CPU to reserve for a container, and containers share unallocated CPU units with other containers on the instance with the same ratio as their allocated amount.")
	c.Int64Var(&cliMemory, "memory", 512, "The number of MiB of memory reserved for the container. If your container attempts to exceed the memory allocated here, the container is killed.")
	c.BoolVar(&cliEssential, "essential", true, "If the essential parameter of a container is marked as true, the failure of that container will stop the task. If the essential parameter of a container is marked as false, then its failure will not affect the rest of the containers in a task. If this parameter is omitted, a container is assumed to be essential.")
	c.StringVar(&cliLinks, "links", "", "A list of links for the container. Each link entry should be in the form of `container_name:alias.`")
	return c
}

func cliRegisterTask(svc *ecs.ECS, args []string) ([]*string, error) {
	var links []*string
	err := cliRegisterTaskParams(args).Parse(args)
	if cliLinks != "" {
		links = aws.StringSlice(strings.Split(cliLinks, ","))
	}
	params := Definition(
		&cliFamily,
		&cliContainerPort,
		&cliHostPort,
		&cliImage,
		&cliCPU,
		&cliMemory,
		cliEssential,
		links)
	resp, err := RegisterTask(svc, params)
	if err != nil {
		return nil, err
	}
	return []*string{resp.TaskDefinition.TaskDefinitionArn}, err
}

// ListTasks Returns a list of tasks for a specified cluster.
func ListTasks(svc *ecs.ECS, params *ecs.ListTasksInput) (*ecs.ListTasksOutput, error) {
	return svc.ListTasks(params)
}

// Returns a `FlagSet` for `ListTasks`
func cliListTasksParams(args []string) *flag.FlagSet {
	var c = cli.Get("", args)
	c.StringVar(&cliClusterName, "cluster", "default", "The short name or full Amazon Resource Name (ARN) of the cluster that hosts the tasks to list. If you do not specify a cluster, the default cluster is assumed..")
	c.StringVar(&cliContainerInstance, "container-instance", "", "The container instance ID or full Amazon Resource Name (ARN) of the container instance with which to filter the list-tasks results. Specifying a containerInstance limits the results to tasks that belong to that container instance.")
	c.StringVar(&cliDesiredStatus, "desired-status", "RUNNING", "The task status that you want to filter the `ListTasks` results with. Specifying a `desiredStatus` of STOPPED will limit the results to tasks that are in the STOPPED status, which can be useful for debugging tasks that are not starting properly or have died or finished. The default status filter is RUNNING.")
	c.StringVar(&cliFamily, "family", "", "The name of the family with which to filter the list-tasks results. Specifying a family limits the results to tasks that belong to that family.")
	c.Int64Var(&cliMaxResults, "max-items", 100, "The maximum number of task results returned by ListTasks in paginated output. When this parameter is used, ListTasks only returns maxResults results in a single page along with a nextToken response element. The remaining results of the initial request can be seen by sending another ListTasks request with the returned nextToken value. This value can be between 1 and 100. If this parameter is not used, then ListTasks returns up to 100 results and a nextToken value if applicable")
	c.StringVar(&cliServiceName, "service-name", "", "The name of the service with which to filter the list-tasks results. Specifying a serviceName limits the results to tasks that belong to that service.")
	return c
}

func cliListTasks(svc *ecs.ECS, args []string) ([]*string, error) {
	err := cliListTasksParams(args).Parse(args)
	params := &ecs.ListTasksInput{
		Cluster:           &cliClusterName,
		ContainerInstance: cli.String(cliContainerInstance),
		DesiredStatus:     cli.String(cliDesiredStatus),
		Family:            cli.String(cliFamily),
		MaxResults:        &cliMaxResults,
		ServiceName:       cli.String(cliServiceName),
	}
	resp, err := ListTasks(svc, params)
	if err != nil {
		return nil, err
	}
	return resp.TaskArns, err
}

// ListTaskDefs Returns a list of task definitions that are registered to your account. You can filter the results by family name with the family parameter or by status with the status parameter.
func ListTaskDefs(svc *ecs.ECS, familyPrefix *string, status *string) (*ecs.ListTaskDefinitionsOutput, error) {
	params := &ecs.ListTaskDefinitionsInput{
		FamilyPrefix: familyPrefix,
		Status:       status,
	}
	resp, err := svc.ListTaskDefinitions(params)
	if err != nil {
		return resp, err
	}
	return resp, nil
}

func cliListTaskDefsParams(args []string) *flag.FlagSet {
	var c = cli.Get("", args)
	c.StringVar(&cliFamily, "family", "", "The full family name with which to filter the list-task-definitions results. Specifying a family limits the listed task definitions to task definition revisions that belong to that family.")
	c.StringVar(&cliStatus, "status", "ACTIVE", "The task definition status with which to filter the list-task-definitions results. By default, only ACTIVE task definitions are listed. By setting this parameter to INACTIVE , you can view task definitions that are INACTIVE as long as an active task or service still references them. Possible values: ACTIVE, INACTIVE")
	return c
}

func cliListTaskDefs(svc *ecs.ECS, args []string) ([]*string, error) {
	err := cliListTaskDefsParams(args).Parse(args)
	resp, err := ListTaskDefs(svc, cli.String(cliFamily), cli.String(cliStatus))
	if err != nil {
		return nil, err
	}
	return resp.TaskDefinitionArns, nil
}

// DescribeTasks Describes an ECS task
func DescribeTasks(svc *ecs.ECS, taskArns []*string, cluster *string) (*ecs.DescribeTasksOutput, error) {
	if len(taskArns) == 0 {
		return nil, errors.New("Tasks can not be blank")
	}
	params := &ecs.DescribeTasksInput{
		Tasks:   taskArns,
		Cluster: cluster,
	}
	resp, err := svc.DescribeTasks(params)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func cliDescribeTasksParams(args []string) *flag.FlagSet {
	var c = cli.Get("", args)
	c.StringVar(&cliClusterName, "cluster", "default", "The short name or full Amazon Resource Name (ARN) of the cluster that hosts the tasks to list. If you do not specify a cluster, the default cluster is assumed..")
	c.StringVar(&cliTasks, "tasks", "", "Comma separated list of taskarns.")
	return c
}

func cliDescribeTasks(svc *ecs.ECS, args []string) ([]*string, error) {
	err := cliDescribeTasksParams(args).Parse(args)
	resp, err := DescribeTasks(
		svc,
		aws.StringSlice(strings.Split(cliTasks, ",")),
		&cliClusterName,
	)
	if err != nil {
		return nil, err
	}
	fail := cli.Failure(resp.Failures, err)
	if fail != nil {
		return nil, fail
	}
	var ret = make([]*string, len(resp.Tasks))
	for k := range resp.Tasks {
		ret[k] = resp.Tasks[k].TaskArn
	}
	return ret, err
}

// StartTask Starts a new task in ECS
func StartTask(svc *ecs.ECS, taskDef *string, containerInstances []*string, cluster *string, startedBy *string, overrides *ecs.TaskOverride) (*ecs.StartTaskOutput, error) {
	params := &ecs.StartTaskInput{
		Cluster:            cluster,
		ContainerInstances: containerInstances,
		TaskDefinition:     taskDef,
		StartedBy:          startedBy,
		Overrides:          overrides,
	}
	resp, err := svc.StartTask(params)
	if err != nil {
		return nil, err
	}
	return resp, err
}

// StartWait Starts a new task a waits until it started successfully.
func StartWait(svc *ecs.ECS, maxTries *int, timeout *int64, taskDef *string, containerInstances []*string, cluster *string, startedBy *string, overrides *ecs.TaskOverride) (*ecs.DescribeTasksOutput, error) {
	tries := 0
	counter := 0
	start, err := StartTask(
		svc,
		taskDef,
		containerInstances,
		cluster,
		startedBy,
		nil,
	)
	if err != nil {
		return nil, err
	}
	fail := cli.Failure(start.Failures, err)
	if fail != nil {
		return nil, fail
	}
	var tasks = make([]*string, len(start.Tasks))
	for i, v := range start.Tasks {
		tasks[i] = v.TaskArn
	}
	for {
		resp, err := DescribeTasks(
			svc,
			tasks,
			&cliClusterName,
		)
		descFail := cli.Failure(resp.Failures, err)
		if descFail != nil {
			return nil, descFail
		}
		for _, v := range resp.Tasks {
			if *v.LastStatus == *v.DesiredStatus {
				counter = counter + 1
			}
			tries = tries + 1
		}
		if counter >= len(resp.Tasks) {
			return resp, nil
		}
		if tries >= *maxTries {
			return resp, errors.New("Max tries (" + strconv.Itoa(*maxTries) + ") reached")
		}
		time.Sleep(time.Duration(*timeout) * time.Second)
	}
}

func cliStartWaitParams(args []string) *flag.FlagSet {
	var c = cli.Get("", args)
	c.Int64Var(&cliTimeout, "timeout", 2, "Wait seconds between two taks polling.")
	c.IntVar(&cliMaxTries, "max-tries", 10, "Max attempts to find a started task.")
	c.StringVar(&cliClusterName, "cluster", "default", "The short name or full Amazon Resource Name (ARN) of the cluster that hosts the tasks to list. If you do not specify a cluster, the default cluster is assumed..")
	c.StringVar(&cliContainerInstance, "container-instances", "", "Comma separated list of container instance IDs or full Amazon Resource Name (ARN) entries for the container instances on which you would like to place your task. The list of container instances to start tasks on is limited to 10.")
	c.StringVar(&cliTaskDef, "task-definition", "", "The family and revision (family:revision ) or full Amazon Resource Name (ARN) of the task definition to start. If a revision is not specified, the latest ACTIVE revision is used.")
	c.StringVar(&cliStartedby, "started-by", "", "An optional tag specified when a task is started. For example if you automatically trigger a task to run a batch process job, you could apply a unique identifier for that job to your task with the startedBy parameter. You can then identify which tasks belong to that job by filtering the results of a list-tasks call with the startedBy value. If a task is started by an Amazon ECS service, then the startedBy parameter contains the deployment ID of the service that starts it.")
	return c
}

func cliStartWait(svc *ecs.ECS, args []string) ([]*string, error) {
	err := cliStartWaitParams(args).Parse(args)
	containerInstances := aws.StringSlice(strings.Split(cliContainerInstance, ","))
	resp, err := StartWait(
		svc,
		&cliMaxTries,
		&cliTimeout,
		&cliTaskDef,
		containerInstances,
		cli.String(cliClusterName),
		cli.String(cliStartedby),
		nil,
	)
	if err != nil {
		return nil, err
	}
	fail := cli.Failure(resp.Failures, err)
	if fail != nil {
		return nil, fail
	}
	var ret = make([]*string, len(resp.Tasks))
	for k := range resp.Tasks {
		ret[k] = resp.Tasks[k].TaskArn
	}
	params := &ecs.DescribeContainerInstancesInput{
		Cluster:            cli.String(cliClusterName),
		ContainerInstances: containerInstances,
	}
	ins, inserr := svc.DescribeContainerInstances(params)
	if inserr != nil {
		return ret, inserr
	}
	insfail := cli.Failure(ins.Failures, err)
	if insfail != nil {
		return nil, insfail
	}
	var ec2Instances = make([]*string, len(ins.ContainerInstances))
	for i, v := range ins.ContainerInstances {
		ec2Instances[i] = v.Ec2InstanceId
	}
	ec2client := ec2.New(nil)
	ec2params := &ec2.DescribeInstancesInput{
		DryRun:      aws.Bool(false),
		InstanceIds: ec2Instances,
	}
	dns, dnserr := ec2client.DescribeInstances(ec2params)
	if dnserr != nil {
		return ret, dnserr
	}
	var result = make([]*string, len(ret)+len(dns.Reservations))
	copy(result, ret)
	for _, r := range dns.Reservations {
		for i, v := range r.Instances {
			result[len(ret)+i] = v.PublicDnsName
		}
	}
	return result, nil
}

func cliStartTaskParams(args []string) *flag.FlagSet {
	var c = cli.Get("", args)
	c.StringVar(&cliClusterName, "cluster", "default", "The short name or full Amazon Resource Name (ARN) of the cluster that hosts the tasks to list. If you do not specify a cluster, the default cluster is assumed..")
	c.StringVar(&cliContainerInstance, "container-instances", "", "Comma separated list of container instance IDs or full Amazon Resource Name (ARN) entries for the container instances on which you would like to place your task. The list of container instances to start tasks on is limited to 10.")
	c.StringVar(&cliTaskDef, "task-definition", "", "The family and revision (family:revision ) or full Amazon Resource Name (ARN) of the task definition to start. If a revision is not specified, the latest ACTIVE revision is used.")
	c.StringVar(&cliStartedby, "started-by", "", "An optional tag specified when a task is started. For example if you automatically trigger a task to run a batch process job, you could apply a unique identifier for that job to your task with the startedBy parameter. You can then identify which tasks belong to that job by filtering the results of a list-tasks call with the startedBy value. If a task is started by an Amazon ECS service, then the startedBy parameter contains the deployment ID of the service that starts it.")
	return c
}

func cliStartTask(svc *ecs.ECS, args []string) ([]*string, error) {
	err := cliStartTaskParams(args).Parse(args)
	resp, err := StartTask(
		svc,
		&cliTaskDef,
		aws.StringSlice(strings.Split(cliContainerInstance, ",")),
		cli.String(cliClusterName),
		cli.String(cliStartedby),
		nil,
	)
	if err != nil {
		return nil, err
	}
	fail := cli.Failure(resp.Failures, err)
	if fail != nil {
		return nil, fail
	}
	var ret = make([]*string, len(resp.Tasks))
	for k := range resp.Tasks {
		ret[k] = resp.Tasks[k].TaskArn
	}
	return ret, nil
}

// StopTask Stops a task
func StopTask(svc *ecs.ECS, task *string, cluster *string) (*ecs.StopTaskOutput, error) {
	params := &ecs.StopTaskInput{
		Task:    task,
		Cluster: cluster,
	}
	resp, err := svc.StopTask(params)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func cliStopTaskParams(args []string) *flag.FlagSet {
	var c = cli.Get("", args)
	c.StringVar(&cliClusterName, "cluster", "default", "The short name or full Amazon Resource Name (ARN) of the cluster that hosts the tasks to list. If you do not specify a cluster, the default cluster is assumed..")
	c.StringVar(&cliTaskDef, "task", "", "The family and revision (family:revision ) or full Amazon Resource Name (ARN) of the task definition to start. If a revision is not specified, the latest ACTIVE revision is used.")
	return c
}

func cliStopTask(svc *ecs.ECS, args []string) ([]*string, error) {
	err := cliStopTaskParams(args).Parse(args)
	resp, err := StopTask(
		svc,
		&cliTaskDef,
		cli.String(cliClusterName),
	)
	if err != nil {
		return nil, err
	}
	return []*string{resp.Task.TaskArn, resp.Task.DesiredStatus, resp.Task.LastStatus}, nil
}

var commands = map[string]cli.Command{
	"desc": {
		cliDescribeTasks,
		"Describes a specified task or tasks.",
		cliDescribeTasksParams,
	},
	"list": {
		cliListTasks,
		"Returns a list of tasks for a specified cluster. You can filter the results by family name, by a particular container instance, or by the desired status of the task with the family , containerInstance , and desiredStatus parameters.",
		cliListTasksParams,
	},
	"definitions": {
		cliListTaskDefs,
		"Returns a list of task definitions that are registered to your account. You can filter the results by family name with the family parameter or by status with the status parameter.",
		cliListTaskDefsParams,
	},
	"register": {
		cliRegisterTask,
		"Registers a new task definition from the supplied family and containerDefinitions.",
		cliRegisterTaskParams,
	},
	"start": {
		cliStartTask,
		"Starts a new task from the specified task definition on the specified container instance or instances. To use the default Amazon ECS scheduler to place your task, use run-task instead.",
		cliStartTaskParams,
	},
	"start-wait": {
		cliStartWait,
		"Starts a new task from the specified task definition on the specified container instance or instances. It's blocks until the specified task starts and print its data.",
		cliStartWaitParams,
	},
	"stop": {
		cliStopTask,
		"Stops a running task.",
		cliStopTaskParams,
	},
}

// Run Main entry point, which runs a command or display a help message.
func Run(command string, args []string) ([]*string, error) {
	return cli.Run(command, commands, args)
}
