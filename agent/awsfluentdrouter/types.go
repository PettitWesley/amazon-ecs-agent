package awsfluentdrouter

import (
	"io"
	"text/template"
)

type AWSFluentdRouterConfig struct {
	Containers  []Container
	ECSMetadata ECSMetadata
}

func (config *AWSFluentdRouterConfig) ToFluentdConfig(wr io.Writer) {
	tmpl := template.Must(template.ParseFiles("agent/awsfluentdrouter/template.conf"))
	tmpl.Execute(wr, *config)
}

type MultilineOptions struct {
	StartRegex string
	EndRegex   string
	Separator  string
}

type ECSMetadata struct {
	Cluster                string
	TaskARN                string
	TaskDefinitionFamily   string
	TaskDefinitionRevision string
}

// Container holds a single container's log configuration
type Container struct {
	Tag              string
	MultilineOptions *MultilineOptions
	Destinations     []Destination
}

// TODO: add validate function- CloudwatchDestination, KinesisStreamsDestination, and KinesisFirehoseDestination are mutually exclusive.
// Only one may be set.
type Destination struct {
	Tag                        string // Same value as Container.tag, duplicated here because of how scope works in go templates
	DestinationUID             string // unique across all destinations for all containers
	FilterOptions              FilterOptions
	CloudwatchDestination      *CloudwatchDestination
	KinesisStreamsDestination  *KinesisStreamsDestination
	KinesisFirehoseDestination *KinesisFirehoseDestination
}

type FilterOptions struct {
	MatchPatterns   []string
	ExcludePatterns []string
}

type CloudwatchDestination struct {
	Region    string
	LogGroup  string
	LogStream string
}

type KinesisFirehoseDestination struct {
	Region             string
	DeliveryStreamName string
}

type KinesisStreamsDestination struct {
	Region     string
	StreamName string
}
