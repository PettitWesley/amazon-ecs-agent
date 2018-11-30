package awsfluentdrouter

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/pkg/errors"
)

const (
	destinationTypeCloudWatchLogs  = "cloudwatch_logs"
	destinationTypeKinesisFirehose = "kinesis_firehose"
	destinationTypeKinesisStreams  = "kinesis_streams"
)

const (
	multilineStartRegexOption = "multiline-start-regexp"
	multilineEndRegexOption   = "multiline-end-regexp"
	multilineSeparatorOption  = "multiline-separator"
)

func NewAWSFluentdRouterConfig(cluster, taskARN, taskDefinitionFamily, taskDefinitionRevision string) *AWSFluentdRouterConfig {
	return &AWSFluentdRouterConfig{
		ECSMetadata: ECSMetadata{
			Cluster:                cluster,
			TaskDefinitionFamily:   taskDefinitionFamily,
			TaskDefinitionRevision: taskDefinitionRevision,
			TaskARN:                taskARN,
		},
	}
}

func (config *AWSFluentdRouterConfig) AddContainer(name string, logOptions map[string]string) error {
	parsedOptions := make(map[string]map[string]string)
	container := Container{
		Tag: name,
	}

	taskID, err := getTaskID(config.ECSMetadata.TaskARN)
	if err != nil {
		return err
	}

	for option, value := range logOptions {
		if strings.HasPrefix(option, "output") {
			err := parseDestinationOption(option, value, parsedOptions)
			if err != nil {
				return err
			}
		} else if strings.HasPrefix(option, "multiline") {
			err := parseMultilineOption(option, value, container)
			if err != nil {
				return err
			}
		}
	}

	for destinationNumber, options := range parsedOptions {
		dest := Destination{
			DestinationUID: fmt.Sprintf("DEST_%s_%s", strings.ToUpper(name), destinationNumber),
			Tag:            name,
		}

		switch destType := options["type"]; destType {
		case destinationTypeCloudWatchLogs:
			cwDest, err := parseCloudWatchDestination(options, name, taskID)
			if err != nil {
				return err
			}
			dest.CloudwatchDestination = cwDest
		case destinationTypeKinesisStreams:
			region, streamName, err := parseKinesisResource(options["destination-arn"])
			if err != nil {
				return err
			}
			dest.KinesisStreamsDestination = &KinesisStreamsDestination{
				Region:     region,
				StreamName: streamName,
			}
		case destinationTypeKinesisFirehose:
			region, deliveryStreamName, err := parseKinesisResource(options["destination-arn"])
			if err != nil {
				return err
			}
			dest.KinesisFirehoseDestination = &KinesisFirehoseDestination{
				Region:             region,
				DeliveryStreamName: deliveryStreamName,
			}
		default:
			return fmt.Errorf("Found unsupported destination type: %s", destType)
		}

		// add filter options
		for option, value := range options {
			if strings.HasPrefix(option, "match") {
				dest.FilterOptions.MatchPatterns = append(dest.FilterOptions.MatchPatterns, value)
			}
			if strings.HasPrefix(option, "exclude") {
				dest.FilterOptions.ExcludePatterns = append(dest.FilterOptions.ExcludePatterns, value)
			}
		}

		container.Destinations = append(container.Destinations, dest)
	}

	config.Containers = append(config.Containers, container)
	return nil
}

func getTaskID(taskARN string) (string, error) {
	ARN, err := arn.Parse(taskARN)
	if err != nil {
		return "", errors.Wrap(err, "Error parsing Task ARN")
	}
	parts := strings.Split(ARN.Resource, "/")
	if parts[0] != "task" {
		return "", fmt.Errorf("Error parsing ECS ARN: Found unexpected resource type: %s", parts[0])
	}
	// whether new or old ECS ARN format is in use, the task ID will be the last piece
	return parts[len(parts)-1], nil
}

func parseKinesisResource(kinesisARN string) (string, string, error) {
	ARN, err := arn.Parse(kinesisARN)
	if err != nil {
		return "", "", errors.Wrap(err, "Error parsing Kinesis ARN")
	}
	parts := strings.Split(ARN.Resource, "/")
	return ARN.Region, parts[1], nil
}

func parseCloudWatchDestination(options map[string]string, containerName, taskID string) (*CloudwatchDestination, error) {
	logARN := options["destination-arn"]
	ARN, err := arn.Parse(logARN)
	if err != nil {
		return nil, errors.Wrap(err, "Error parsing CloudWatch Logs ARN")
	}
	logGroup, logStream, err := parseLogResource(ARN.Resource, containerName, taskID)
	if err != nil {
		return nil, err
	}

	cwDest := &CloudwatchDestination{
		Region:    ARN.Region,
		LogGroup:  logGroup,
		LogStream: logStream,
	}

	return cwDest, nil
}

func parseLogResource(resource string, containerName, taskID string) (string, string, error) {
	parts := strings.Split(resource, ":")
	if parts[0] != "log-group" {
		return "", "", fmt.Errorf("Error parsing CloudWatch ARN: Found unexpected resource type: %s", parts[0])
	}
	logGroup := parts[1]
	logStream := parts[2]
	// prefix* => prefix/container_name/task_id
	if strings.HasSuffix(logStream, "*") {
		logStream = strings.Replace(logStream, "*", fmt.Sprintf("/%s/%s", containerName, taskID), 1)
		if strings.HasPrefix(logStream, "/") {
			logStream = strings.Replace(logStream, "/", "", 1)
		}
	}

	return logGroup, logStream, nil
}

func parseMultilineOption(option, value string, container Container) error {
	if container.MultilineOptions == nil {
		container.MultilineOptions = &MultilineOptions{}
	}
	if option == multilineStartRegexOption {
		container.MultilineOptions.StartRegex = value
	} else if option == multilineEndRegexOption {
		container.MultilineOptions.EndRegex = value
	} else if option == multilineSeparatorOption {
		container.MultilineOptions.Separator = value
	} else {
		return fmt.Errorf("Could not parse log option %s", option)
	}
	return nil
}

func parseDestinationOption(option, value string, parsedOptions map[string]map[string]string) error {
	r := regexp.MustCompile(`^output(\d{1})-([-\w]+)`)
	match := r.FindStringSubmatch(option)
	if len(match) < 3 {
		return fmt.Errorf("Could not parse log option %s", option)
	}
	destinationNumber := match[1]
	destinationOption := match[2]
	if _, exists := parsedOptions[destinationNumber]; !exists {
		parsedOptions[destinationNumber] = make(map[string]string)
	}
	parsedOptions[destinationNumber][destinationOption] = value
	return nil
}
