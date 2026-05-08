package web

import "embed"

// FS contains server-rendered templates and static assets. Building the Go
// binary is enough to ship the runnable application.
//
//go:embed templates/* public/*
var FS embed.FS
