package awsfluentdrouter

var fluentdConfigTemplate = `
<source>
  @type unix
  path /socket/fluentd.sock
  @id docker_input
</source>

<filter **>
  @type record_transformer
  <record>
    ecs_cluster {{.ECSMetadata.Cluster}}
    ecs_task_arn {{.ECSMetadata.TaskARN}}
    ecs_task_definition {{.ECSMetadata.TaskDefinitionFamily}}:{{.ECSMetadata.TaskDefinitionRevision}}
  </record>
</filter>

{{range .Containers}}
{{if .MultilineOptions}}
<filter {{.Tag}}>
  @type concat
  key log
  {{if .MultilineOptions.StartRegex}}
  multiline_start_regexp {{.MultilineOptions.StartRegex}}
  {{end}}
  {{if .MultilineOptions.EndRegex}}
  multiline_end_regexp {{.MultilineOptions.EndRegex}}
  {{end}}
  {{if .MultilineOptions.Separator}}
  separator {{.MultilineOptions.Separator}}
  {{end}}
</filter>
{{end}}
<match {{.Tag}}>
  @type copy
  {{range .Destinations}}
  <store>
    @type relabel
    @label {{.DestinationUID}}
  </store>
  {{end}}
</match>
{{end}}

{{range .Containers}}
{{range .Destinations}}
<label @{{.DestinationUID}}>
  <filter {{.Tag}}>
    @type grep
    {{range .FilterOptions.MatchPatterns}}
    <regexp>
      key log
      pattern {{.}}
    </regexp>
    {{end}}
    {{range .FilterOptions.ExcludePatterns}}
    <exclude>
      key log
      pattern {{.}}
    </exclude>
    {{end}}
  </filter>

  {{if .KinesisFirehoseDestination}}
  <match {{.Tag}}>
    @type kinesis_firehose
    @id   {{.DestinationUID}}_OUTPUT
    region {{.KinesisFirehoseDestination.Region}}
    delivery_stream_name {{.KinesisFirehoseDestination.DeliveryStreamName}}
  </match>
  {{end}}
  {{if .KinesisStreamsDestination}}
  <match {{.Tag}}>
    @type kinesis_firehose
    @id   {{.DestinationUID}}_OUTPUT
    region {{.KinesisStreamsDestination.Region}}
    stream_name {{.KinesisStreamsDestination.StreamName}}
  </match>
  {{end}}
  {{if .CloudwatchDestination}}
  <match {{.Tag}}>
    @type cloudwatch_logs
    @id   {{.DestinationUID}}_OUTPUT
    region {{.CloudwatchDestination.Region}}
    log_group_name {{.CloudwatchDestination.LogGroup}}
    log_stream_name {{.CloudwatchDestination.LogStream}}
    auto_create_stream true
  </match>
  {{end}}
</label>
{{end}}
{{end}}
`
