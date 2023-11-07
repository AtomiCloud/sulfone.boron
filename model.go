package main

import "github.com/AtomiCloud/sulfone.boron/docker_executor"

type StartExecutorReq struct {
	SessionId         string                                `json:"session_id"`
	Template          docker_executor.TemplateVersionRes    `json:"template"`
	WriteVolReference docker_executor.DockerVolumeReference `json:"write_vol_reference"`
	Merger            docker_executor.MergerReq             `json:"merger"`
}

type ProblemDetails struct {
	Title   string      `json:"title"`
	Status  int         `json:"status"`
	Detail  string      `json:"detail"`
	Type    string      `json:"type"`
	TraceId *string     `json:"trace_id"`
	Data    interface{} `json:"data"`
}
