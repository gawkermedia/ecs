package cluster

import (
	"flag"

	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/gawkermedia/ecs/cli"
)

// CLI params
var cliMaxResults int64
var cliClusterName string

// CLI params END

// ListClusters List ECS clusters
func ListClusters(svc *ecs.ECS, maxResults *int64) (*ecs.ListClustersOutput, error) {
	params := &ecs.ListClustersInput{
		MaxResults: maxResults,
	}
	resp, err := svc.ListClusters(params)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func cliListClustersParams(args []string) *flag.FlagSet {
	var c = cli.Get("", args)
	c.Int64Var(&cliMaxResults, "max-items", 1, "The maximum number of cluster results returned by ListClusters in paginated output")
	return c
}

func cliListClusters(svc *ecs.ECS, args []string) ([]*string, error) {
	err := cliListClustersParams(args).Parse(args)
	resp, err := ListClusters(svc, &cliMaxResults)
	if err != nil {
		return nil, err
	}
	return resp.ClusterArns, nil
}

// CreateCluster Creates a new ECS cluster
func CreateCluster(svc *ecs.ECS, name *string) (*ecs.CreateClusterOutput, error) {
	params := &ecs.CreateClusterInput{
		ClusterName: name,
	}
	resp, err := svc.CreateCluster(params)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func cliClusterNameParams(args []string) *flag.FlagSet {
	var c = cli.Get("", args)
	c.StringVar(&cliClusterName, "cluster", "defaut", "The name of your cluster. If you do not specify a name for your cluster, If you do not specify a cluster, the default cluster is assumed. Up to 255 letters (uppercase and lowercase), numbers, hyphens, and underscores are allowed.")
	return c
}

func cliCreateCluster(svc *ecs.ECS, args []string) ([]*string, error) {
	err := cliClusterNameParams(args).Parse(args)
	resp, err := CreateCluster(svc, &cliClusterName)
	if err != nil {
		return nil, err
	}
	return []*string{resp.Cluster.ClusterName}, nil
}

// DeleteCluster Removes an ECS cluster
func DeleteCluster(svc *ecs.ECS, name *string) (*ecs.DeleteClusterOutput, error) {
	params := &ecs.DeleteClusterInput{
		Cluster: name,
	}
	resp, err := svc.DeleteCluster(params)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func cliDeleteCluster(svc *ecs.ECS, args []string) ([]*string, error) {
	cliClusterNameParams(args)
	resp, err := DeleteCluster(svc, &cliClusterName)
	if err != nil {
		return nil, err
	}
	return []*string{resp.Cluster.ClusterName}, nil
}

var commands = map[string]cli.Command{
	"list": {
		cliListClusters,
		"Returns a list of existing clusters.",
		cliListClustersParams,
	},
	"create": {
		cliCreateCluster,
		"Creates a new Amazon ECS cluster.",
		cliClusterNameParams,
	},
	"delete": {
		cliDeleteCluster,
		"Deletes the specified cluster. You must deregister all container instances from this cluster before you may delete it.",
		cliClusterNameParams,
	},
}

// Run Main entry point, which runs a command or display a help message.
func Run(command string, args []string) ([]*string, error) {
	return cli.Run(command, commands, args)
}
