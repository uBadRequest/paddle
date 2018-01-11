package pipeline

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"
)

type PodSecret struct {
	Name  string
	Store string
	Key   string
}

type PodDefinition struct {
	PodName     string
	StepName    string
	StepVersion string
	BranchName  string
	Namespace   string
	Bucket      string
	Secrets     []PodSecret

	Step PipelineDefinitionStep
}

const podTemplate = `
apiVersion: v1
kind: Pod
metadata:
  name: "{{ .PodName }}"
  namespace: {{ .Namespace }}
  labels:
    canoe.executor: paddle
    canoe.step.name: {{ .StepName }}
    canoe.step.branch: {{ .BranchName }}
    canoe.step.version: {{ .StepVersion }}
spec:
  restartPolicy: Never
  volumes:
    -
      name: shared-data
      emptyDir:
        medium: ''
  containers:
    -
      name: main
      image: "{{ .Step.Image }}"
      volumeMounts:
        -
          name: shared-data
          mountPath: /data
      resources:
        limits:
          cpu: "{{ .Step.Resources.CPU }}"
          memory: "{{ .Step.Resources.Memory }}"
      command:
        - "/bin/sh"
        - "-c"
        - "while true; do if [ -e /data/first-step.txt ]; then ((
          {{ range $index, $command := .Step.Commands }}
          ({{ $command }}) &&
          {{ end }}
          touch /data/main-passed.txt) || (touch /data/main-failed.txt && exit 1)) && touch /data/main.txt; break; fi; done"
      env:
        -
          name: INPUT_PATH
          value: /data/input
        -
          name: OUTPUT_PATH
          value: /data/output
        -
          name: AWS_ACCESS_KEY_ID
          valueFrom:
            secretKeyRef:
              name: aws-credentials-training
              key: aws-access-key-id
        -
          name: AWS_SECRET_ACCESS_KEY
          valueFrom:
            secretKeyRef:
              name: aws-credentials-training
              key: aws-secret-access-key
        {{ range $index, $secret := .Secrets }}
        -
          name: {{ $secret.Name }}
          valueFrom:
            secretKeyRef:
              name: {{ $secret.Store }}
              key: {{ $secret.Key }}
        {{ end }}
    -
      name: paddle
      image: "219541440308.dkr.ecr.eu-west-1.amazonaws.com/paddlecontainer:latest"
      volumeMounts:
        -
          name: shared-data
          mountPath: /data
      command:
        - "/bin/sh"
        - "-c"
        - "mkdir -p $INPUT_PATH $OUTPUT_PATH &&
          {{ range $index, $input := .Step.Inputs }}
          paddle data get {{ $input.Step }}/{{ $input.Version }} $INPUT_PATH -b {{ $input.Branch | sanitizeName }} -p {{ $input.Path }} &&
          {{ end }}
          touch /data/first-step.txt &&
          echo first step finished &&
          (while true; do
            if [ -e /data/main-failed.txt ]; then
              exit 1;
            fi;
            if [ -e /data/main-passed.txt ]; then
              paddle data commit $OUTPUT_PATH {{ .StepName }}/{{ .Step.Version }} -b {{ .BranchName }};
              exit 0;
            fi;
          done)"
      env:
        -
          name: BUCKET
          value: "{{ .Bucket }}"
        -
          name: AWS_REGION
          value: eu-west-1
        -
          name: INPUT_PATH
          value: /data/input
        -
          name: OUTPUT_PATH
          value: /data/output
        -
          name: AWS_ACCESS_KEY_ID
          valueFrom:
            secretKeyRef:
              name: aws-credentials
              key: aws-access-key-id
        -
          name: AWS_SECRET_ACCESS_KEY
          valueFrom:
            secretKeyRef:
              name: aws-credentials
              key: aws-secret-access-key
        {{ range $index, $secret := .Secrets }}
        -
          name: {{ $secret.Name }}
          valueFrom:
            secretKeyRef:
              name: {{ $secret.Store }}
              key: {{ $secret.Key }}
        {{ end }}
`

func NewPodDefinition(pipelineDefinition *PipelineDefinition, pipelineDefinitionStep *PipelineDefinitionStep) *PodDefinition {
	stepName := sanitizeName(pipelineDefinitionStep.Step)
	branchName := sanitizeName(pipelineDefinitionStep.Branch)
	stepVersion := sanitizeName(pipelineDefinitionStep.Version)
	podName := fmt.Sprintf("%s-%s-%s", sanitizeName(pipelineDefinition.Pipeline), stepName, branchName)

	return &PodDefinition{
		PodName:     podName,
		Namespace:   pipelineDefinition.Namespace,
		Step:        *pipelineDefinitionStep,
		Bucket:      pipelineDefinition.Bucket,
		StepName:    stepName,
		StepVersion: stepVersion,
		BranchName:  branchName,
		Secrets:     []PodSecret{},
	}

}

func (p PodDefinition) compile() *bytes.Buffer {
	fmap := template.FuncMap{
		"sanitizeName": sanitizeName,
	}
	tmpl := template.Must(template.New("podTemplate").Funcs(fmap).Parse(podTemplate))
	buffer := new(bytes.Buffer)
	err := tmpl.Execute(buffer, p)
	if err != nil {
		panic(err.Error())
	}
	return buffer
}

func sanitizeName(name string) string {
	str := strings.ToLower(name)
	str = strings.Replace(str, "_", "-", -1)
	str = strings.Replace(str, "/", "-", -1)
	return str
}
