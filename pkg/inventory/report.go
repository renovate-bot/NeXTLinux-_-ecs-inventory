package inventory

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/aws/aws-sdk-go/service/ecs/ecsiface"

	"github.com/anchore/ecs-inventory/internal/logger"
	"github.com/anchore/ecs-inventory/pkg/connection"
	"github.com/anchore/ecs-inventory/pkg/reporter"
)

func reportToStdout(report reporter.Report) error {
	enc := json.NewEncoder(os.Stdout)
	// prevent > and < from being escaped in the payload
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(report); err != nil {
		return fmt.Errorf("unable to show inventory: %w", err)
	}
	return nil
}

func HandleReport(report reporter.Report, anchoreDetails connection.AnchoreInfo, quiet, dryRun bool) error {
	switch {
	case dryRun:
		logger.Log.Info("Dry run specified, not reporting inventory")
	case anchoreDetails.IsValid():
		if err := reporter.Post(report, anchoreDetails); err != nil {
			return fmt.Errorf("unable to report Inventory to Anchore: %w", err)
		}
	default:
		logger.Log.Debug("Anchore details not specified, not reporting inventory")
	}

	if !quiet {
		return reportToStdout(report)
	}
	return nil
}

func GetInventoryReportsForRegion(region string, anchoreDetails connection.AnchoreInfo, quiet, dryRun, metadata bool) error {
	logger.Log.Info("Getting Inventory Reports for region", "region", region)
	sessConfig := &aws.Config{}
	if region != "" {
		sessConfig.Region = aws.String(region)
	}
	sess, err := session.NewSession(sessConfig)
	if err != nil {
		logger.Log.Error("Failed to create AWS session", err)
	}

	err = checkAWSCredentials(sess)
	if err != nil {
		return err
	}

	ecsClient := ecs.New(sess)

	clusters, err := fetchClusters(ecsClient)
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	wg.Add(len(clusters))

	for _, cluster := range clusters {
		go func(cluster string) {
			defer wg.Done()

			ecsClient := ecs.New(sess)

			report, err := GetInventoryReportForCluster(cluster, ecsClient, metadata)
			if err != nil {
				logger.Log.Error("Failed to get inventory report for cluster", err)
			}

			// Only report if there are containers present in the cluster
			if len(report.Containers) != 0 {
				err = HandleReport(report, anchoreDetails, quiet, dryRun)
				if err != nil {
					logger.Log.Error("Failed to report inventory for cluster", err)
				}
			}
		}(*cluster)
	}

	wg.Wait()
	return nil
}

// GetInventoryReportForCluster is an atomic method for getting in-use image results, for a cluster
func GetInventoryReportForCluster(cluster string, ecsClient ecsiface.ECSAPI, metadata bool) (reporter.Report, error) {
	logger.Log.Debug("Found cluster", "cluster", cluster)

	report := reporter.Report{
		Timestamp:   time.Now().UTC().Format(time.RFC3339),
		ClusterName: cluster,
	}
	tasks, err := fetchTasksFromCluster(ecsClient, cluster)
	if err != nil {
		return reporter.Report{}, err
	}

	if metadata {
		servicesMeta := []reporter.Service{}
		services, err := fetchServicesFromCluster(ecsClient, cluster)
		if err != nil {
			return reporter.Report{}, err
		}
		if len(services) == 0 {
			logger.Log.Debug("No services found in cluster", "cluster", cluster)
		} else {
			servicesMeta, err = fetchServicesMetadata(ecsClient, cluster, services)
			if err != nil {
				return reporter.Report{}, err
			}
		}
		report.Services = servicesMeta
	}

	// Must be at least one task to continue
	if len(tasks) == 0 {
		logger.Log.Debug("No tasks found in cluster", "cluster", cluster)
	} else {
		logger.Log.Debug("Found tasks in cluster", "cluster", cluster, "taskCount", len(tasks))

		if metadata {
			taskMeta, err := fetchTasksMetadata(ecsClient, cluster, tasks)
			if err != nil {
				return reporter.Report{}, err
			}
			report.Tasks = taskMeta
		}

		containers, err := fetchContainersFromTasks(ecsClient, cluster, tasks)
		if err != nil {
			return reporter.Report{}, err
		}
		report.Containers = containers
		logger.Log.Info("Found containers in cluster", "cluster", cluster, "containerCount", len(containers))
	}

	return report, nil
}
