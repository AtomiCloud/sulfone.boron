package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	rt "runtime"

	"github.com/AtomiCloud/sulfone.boron/docker_executor"
	"github.com/docker/docker/client"
	"github.com/gin-gonic/gin"
)

func stringifyErrors(e []error) []string {
	var errs []string
	for _, err := range e {
		errs = append(errs, err.Error())
	}
	return errs
}

func server(registryEndpoint string) {
	r := gin.Default()
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, docker_executor.StandardResponse{Status: "OK"})
	})

	r.DELETE("/executor/:sessionId", func(ctx *gin.Context) {
		sessionId := ctx.Param("sessionId")
		dCli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
		if err != nil {
			panic(err)
		}
		defer func(dCli *client.Client) {
			_ = dCli.Close()
		}(dCli)
		cpu := rt.NumCPU()
		d := docker_executor.DockerClient{
			Docker:           dCli,
			Context:          ctx,
			ParallelismLimit: cpu,
		}
		exec := docker_executor.Executor{
			Docker:   d,
			Template: docker_executor.TemplateVersionRes{},
		}
		e := exec.Clean(sessionId)
		if len(e) > 0 {
			ctx.JSON(http.StatusBadRequest, ProblemDetails{
				Title:   "Failed to clean",
				Status:  400,
				Detail:  "Failed to session " + sessionId,
				Type:    "https://developer.mozilla.org/en-US/docs/Web/HTTP/Status/400",
				TraceId: nil,
				Data:    e,
			})
			return
		}
		ctx.JSON(200, docker_executor.StandardResponse{Status: "OK"})
	})

	r.POST("/executor/:sessionId", func(ctx *gin.Context) {
		sessionId := ctx.Param("sessionId")
		cpu := rt.NumCPU()

		// req
		var req docker_executor.BuildReq
		err := ctx.BindJSON(&req)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, ProblemDetails{
				Title:   "Failed to clean",
				Status:  400,
				Detail:  "Failed to session " + sessionId,
				Type:    "https://developer.mozilla.org/en-US/docs/Web/HTTP/Status/400",
				TraceId: nil,
				Data:    []string{err.Error()},
			})
			return
		}
		merger := docker_executor.Merger{
			ParallelismLimit: cpu,
			RegistryClient: docker_executor.RegistryClient{
				Endpoint: registryEndpoint,
			},
			Template:  req.Template,
			SessionId: sessionId,
		}
		mergePath, errs := merger.Merge(req)
		if len(errs) > 0 {
			ctx.JSON(http.StatusBadRequest, ProblemDetails{
				Title:   "Failed to clean",
				Status:  400,
				Detail:  "Failed to session " + sessionId,
				Type:    "https://developer.mozilla.org/en-US/docs/Web/HTTP/Status/400",
				TraceId: nil,
				Data:    stringifyErrors(errs),
			})
			return
		}
		// zip

		c := docker_executor.DockerContainerReference{
			CyanId:    req.MergerId,
			CyanType:  "merger",
			SessionId: sessionId,
		}
		ep := docker_executor.DockerContainerToString(c)
		endpoint := "http://" + ep + ":9000/zip"

		zipR := docker_executor.ZipReq{
			TargetDir: mergePath,
		}
		jsonValue, err := json.Marshal(zipR)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, ProblemDetails{
				Title:   "Error encoding JSON",
				Status:  400,
				Detail:  "Failed encode JSON zipping request",
				Type:    "https://developer.mozilla.org/en-US/docs/Web/HTTP/Status/400",
				TraceId: nil,
				Data:    []string{err.Error()},
			})
			return
		}
		jsonBody := bytes.NewReader(jsonValue)

		zipReq, err := http.NewRequest("POST", endpoint, jsonBody)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, ProblemDetails{
				Title:   "Failed to generate upstream request",
				Status:  400,
				Detail:  "http.NewRequest return error when generating request for upstream errors",
				Type:    "https://developer.mozilla.org/en-US/docs/Web/HTTP/Status/400",
				TraceId: nil,
				Data:    []string{err.Error()},
			})
			return
		}

		zipReq.Header.Set("Content-Type", "application/json")
		cl := &http.Client{}
		resp, err := cl.Do(zipReq)
		if err != nil {
			ctx.JSON(http.StatusServiceUnavailable, ProblemDetails{
				Title:   "Failed to contract upstream server",
				Status:  503,
				Detail:  "Error contacting upstream (merger) server for zipping",
				Type:    "https://developer.mozilla.org/en-US/docs/Web/HTTP/Status/503",
				TraceId: nil,
				Data:    []string{err.Error()},
			})
			return
		}
		defer func(Body io.ReadCloser) {
			_ = Body.Close()
		}(resp.Body)

		for key, values := range resp.Header {
			for _, value := range values {
				ctx.Header(key, value)
			}
		}
		_, err = io.Copy(ctx.Writer, resp.Body)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, ProblemDetails{
				Title:   "Failed to generate streaming response",
				Status:  400,
				Detail:  "Error copying upstream stream zip response as response of coordinator",
				Type:    "https://developer.mozilla.org/en-US/docs/Web/HTTP/Status/400",
				TraceId: nil,
				Data:    []string{err.Error()},
			})
			return
		}
	})

	r.POST("/executor", func(ctx *gin.Context) {
		var req StartExecutorReq
		err := ctx.BindJSON(&req)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, ProblemDetails{
				Title:   "Failed to bind to request",
				Status:  400,
				Detail:  "Request Body JSON does not match StartExecutorReq",
				Type:    "https://developer.mozilla.org/en-US/docs/Web/HTTP/Status/400",
				TraceId: nil,
				Data:    []string{err.Error()},
			})
			return
		}
		dCli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
		if err != nil {
			panic(err)
		}
		defer func(dCli *client.Client) {
			_ = dCli.Close()
		}(dCli)
		cpu := rt.NumCPU()
		d := docker_executor.DockerClient{
			Docker:           dCli,
			Context:          ctx,
			ParallelismLimit: cpu,
		}
		exec := docker_executor.Executor{
			Docker:   d,
			Template: req.Template,
		}
		err = d.EnforceNetwork()
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, ProblemDetails{
				Title:   "Failed to configure network",
				Status:  503,
				Detail:  "Failed to start cyanprint Docker bridge network",
				Type:    "https://developer.mozilla.org/en-US/docs/Web/HTTP/Status/503",
				TraceId: nil,
				Data:    []string{err.Error()},
			})
			return
		}
		readVolRef := docker_executor.DockerVolumeReference{
			CyanId:    req.Template.Principal.ID,
			SessionId: "",
		}

		errs := exec.Start(req.SessionId, readVolRef, req.WriteVolReference, req.Merger)
		if len(errs) > 0 {
			ctx.JSON(http.StatusInternalServerError, ProblemDetails{
				Title:   "Failed to start executor",
				Status:  503,
				Detail:  "Failed to start cyanprint executor",
				Type:    "https://developer.mozilla.org/en-US/docs/Web/HTTP/Status/503",
				TraceId: nil,
				Data:    stringifyErrors(errs),
			})
		} else {
			ctx.JSON(http.StatusOK, docker_executor.StandardResponse{
				Status: "OK",
			})
		}
	})

	r.POST("/executor/:sessionId/warm", func(ctx *gin.Context) {

		sessionId := ctx.Param("sessionId")
		var template docker_executor.TemplateVersionRes
		err := ctx.BindJSON(&template)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, ProblemDetails{
				Title:   "Failed to bind to request",
				Status:  400,
				Detail:  "Request Body JSON does not match TemplateVersionRes",
				Type:    "https://developer.mozilla.org/en-US/docs/Web/HTTP/Status/400",
				TraceId: nil,
				Data:    []string{err.Error()},
			})
			return
		}
		dCli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
		if err != nil {
			panic(err)
		}
		defer func(dCli *client.Client) {
			_ = dCli.Close()
		}(dCli)
		cpu := rt.NumCPU()
		d := docker_executor.DockerClient{
			Docker:           dCli,
			Context:          ctx,
			ParallelismLimit: cpu,
		}
		exec := docker_executor.Executor{
			Docker:   d,
			Template: template,
		}
		err = d.EnforceNetwork()
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, ProblemDetails{
				Title:   "Failed to configure network",
				Status:  503,
				Detail:  "Failed to start cyanprint Docker bridge network",
				Type:    "https://developer.mozilla.org/en-US/docs/Web/HTTP/Status/503",
				TraceId: nil,
				Data:    []string{err.Error()},
			})
			return
		}
		session, volRef, errs := exec.Warm(sessionId)
		if len(errs) > 0 {
			ctx.JSON(http.StatusBadRequest, ProblemDetails{
				Title:   "Failed to warm executor",
				Status:  400,
				Detail:  "Failed to warn executor image, templates, and volumes",
				Type:    "https://developer.mozilla.org/en-US/docs/Web/HTTP/Status/400",
				TraceId: nil,
				Data:    stringifyErrors(errs),
			})
		}

		ctx.JSON(http.StatusOK, gin.H{
			"session_id": session,
			"vol_ref":    volRef,
		})

	})

	r.POST("/template/warm", func(ctx *gin.Context) {
		var template docker_executor.TemplateVersionRes
		err := ctx.BindJSON(&template)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, ProblemDetails{
				Title:   "Failed to bind to request",
				Status:  400,
				Detail:  "Request Body JSON does not match TemplateVersionRes",
				Type:    "https://developer.mozilla.org/en-US/docs/Web/HTTP/Status/400",
				TraceId: nil,
				Data:    []string{err.Error()},
			})
			return
		}
		dCli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
		if err != nil {
			panic(err)
		}
		defer func(dCli *client.Client) {
			_ = dCli.Close()
		}(dCli)
		cpu := rt.NumCPU()
		d := docker_executor.DockerClient{
			Docker:           dCli,
			Context:          ctx,
			ParallelismLimit: cpu,
		}
		exec := docker_executor.TemplateExecutor{
			Docker:   d,
			Template: template.Principal,
		}
		err = d.EnforceNetwork()
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, ProblemDetails{
				Title:   "Failed to configure network",
				Status:  503,
				Detail:  "Failed to start cyanprint Docker bridge network",
				Type:    "https://developer.mozilla.org/en-US/docs/Web/HTTP/Status/503",
				TraceId: nil,
				Data:    []string{err.Error()},
			})
			return
		}
		errs := exec.WarmTemplate()
		if len(errs) > 0 {
			ctx.JSON(http.StatusBadRequest, ProblemDetails{
				Title:   "Failed to warm template",
				Status:  400,
				Detail:  "Failed to warn template image, templates, and volumes",
				Type:    "https://developer.mozilla.org/en-US/docs/Web/HTTP/Status/400",
				TraceId: nil,
				Data:    stringifyErrors(errs),
			})
		}
		ctx.JSON(http.StatusOK, docker_executor.StandardResponse{Status: "OK"})

	})

	// proxy
	r.POST("/proxy/template/:cyanId/api/template/init", func(c *gin.Context) {

		cyanId := c.Param("cyanId")

		fmt.Println("📇 Cyan ID:", cyanId)
		d := docker_executor.DockerContainerReference{
			CyanId:    cyanId,
			CyanType:  "template",
			SessionId: "",
		}
		endpoint := "http://" + docker_executor.DockerContainerToString(d) + ":5550/api/template/init"
		fmt.Println("🌐 Upstream Endpoint:", endpoint)
		fmt.Println("🆕 Start forwarding request...")

		reqBody, err := io.ReadAll(c.Request.Body)
		if err != nil {
			// Handle error
			c.JSON(http.StatusBadGateway, ProblemDetails{
				Title:   "Read request failed",
				Status:  400,
				Detail:  "Failed read the initial request body",
				Type:    "https://developer.mozilla.org/en-US/docs/Web/HTTP/Status/400",
				TraceId: nil,
				Data:    []string{err.Error()},
			})
			return
		}

		fmt.Println("📦 Request Body:", string(reqBody))

		resp, err := http.Post(endpoint, c.GetHeader("Content-Type"), bytes.NewBuffer(reqBody))
		if err != nil {
			// Handle error
			c.JSON(http.StatusBadGateway, ProblemDetails{
				Title:   "Upstream failed",
				Status:  502,
				Detail:  "Failed to forward request to upstream template",
				Type:    "https://developer.mozilla.org/en-US/docs/Web/HTTP/Status/502",
				TraceId: nil,
				Data:    []string{err.Error()},
			})
			return
		}
		defer func(Body io.ReadCloser) {
			_ = Body.Close()
		}(resp.Body)
		fmt.Println("✅ Request forwarded successfully")
		// Read the response from the new endpoint
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			// Handle error
			// Handle error
			c.JSON(http.StatusBadGateway, ProblemDetails{
				Title:   "Upstream failed",
				Status:  502,
				Detail:  "Failed to read respond from upstream template",
				Type:    "https://developer.mozilla.org/en-US/docs/Web/HTTP/Status/502",
				TraceId: nil,
				Data:    []string{err.Error()},
			})
			return
		}
		fmt.Println("Status Code from upstream:", resp.StatusCode)
		c.Data(resp.StatusCode, "Content-Type", body)

	})
	r.POST("/proxy/template/:cyanId/api/template/validate", func(c *gin.Context) {

		cyanId := c.Param("cyanId")
		fmt.Println("📇 Cyan ID:", cyanId)

		d := docker_executor.DockerContainerReference{
			CyanId:    cyanId,
			CyanType:  "template",
			SessionId: "",
		}
		endpoint := "http://" + docker_executor.DockerContainerToString(d) + ":5550/api/template/validate"
		fmt.Println("🌐 Upstream Endpoint:", endpoint)
		fmt.Println("🆕 Start forwarding request...")

		reqBody, err := io.ReadAll(c.Request.Body)
		if err != nil {
			// Handle error
			c.JSON(http.StatusBadGateway, ProblemDetails{
				Title:   "Read request failed",
				Status:  400,
				Detail:  "Failed read the initial request body",
				Type:    "https://developer.mozilla.org/en-US/docs/Web/HTTP/Status/400",
				TraceId: nil,
				Data:    []string{err.Error()},
			})
			return
		}

		fmt.Println("📦 Request Body:", string(reqBody))

		// Forward the request body directly without reading it first
		resp, err := http.Post(endpoint, c.GetHeader("Content-Type"), bytes.NewBuffer(reqBody))
		if err != nil {
			// Handle error
			c.JSON(http.StatusBadGateway, ProblemDetails{
				Title:   "Upstream failed",
				Status:  502,
				Detail:  "Failed to forward request to upstream template",
				Type:    "https://developer.mozilla.org/en-US/docs/Web/HTTP/Status/502",
				TraceId: nil,
				Data:    []string{err.Error()},
			})
			return
		}
		defer func(Body io.ReadCloser) {
			_ = Body.Close()
		}(resp.Body)

		// Read the response from the new endpoint
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			// Handle error
			c.JSON(http.StatusBadGateway, ProblemDetails{
				Title:   "Upstream failed",
				Status:  502,
				Detail:  "Failed to read respond from upstream template",
				Type:    "https://developer.mozilla.org/en-US/docs/Web/HTTP/Status/502",
				TraceId: nil,
				Data:    []string{err.Error()},
			})
			return
		}
		c.Data(resp.StatusCode, "Content-Type", body)
	})

	// for merger
	r.POST("/merge/:sessionId", func(c *gin.Context) {
		sessionId := c.Param("sessionId")

		cpu := rt.NumCPU()

		var req docker_executor.MergeReq
		err := c.BindJSON(&req)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"errors": []string{err.Error()}})
			return
		}

		m := docker_executor.Merger{
			ParallelismLimit: cpu,
			RegistryClient: docker_executor.RegistryClient{
				Endpoint: registryEndpoint,
			},
			Template:  req.Template,
			SessionId: sessionId,
		}
		err = m.MergeFiles(req.FromDirs, req.ToDir)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": []string{err.Error()}})
			return
		}
		c.JSON(http.StatusOK, docker_executor.StandardResponse{Status: "OK"})
	})

	r.POST("/zip", func(c *gin.Context) {
		var req docker_executor.ZipReq
		err := c.BindJSON(&req)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"errors": []string{err.Error()}})
			return
		}
		pr, pw := io.Pipe()

		// Use a goroutine to stream the tar archive
		go func() {
			defer func(pw *io.PipeWriter) {
				_ = pw.Close()
			}(pw)
			gw := gzip.NewWriter(pw)
			defer func(gw *gzip.Writer) {
				_ = gw.Close()
			}(gw)
			tw := tar.NewWriter(gw)
			defer func(tw *tar.Writer) {
				_ = tw.Close()
			}(tw)

			// Your directory to tar and zip
			dir := req.TargetDir

			// Walk through every file in the folder
			_ = filepath.Walk(dir, func(file string, fi os.FileInfo, err error) error {
				// Return on any error
				if err != nil {
					return err
				}

				if file == dir {
					return nil
				}

				// Create a new dir/file header
				header, err := tar.FileInfoHeader(fi, fi.Name())
				if err != nil {
					return err
				}

				relPath, err := filepath.Rel(dir, file)
				if err != nil {
					return err
				}
				// Update the name to correctly reflect the desired directory structure
				header.Name = relPath

				// Write the header
				if errr := tw.WriteHeader(header); errr != nil {
					return errr
				}

				// If not a dir, write file content
				if !fi.Mode().IsDir() {
					data, e := os.Open(file)
					if e != nil {
						return e
					}
					defer func(data *os.File) {
						_ = data.Close()
					}(data)
					if _, er := io.Copy(tw, data); er != nil {
						return er
					}
				}
				return nil
			})
		}()

		// Set the header and serve the file
		c.Header("Content-Description", "File Transfer")
		c.Header("Content-Type", "application/octet-stream")
		c.Header("Content-Disposition", "attachment; filename=cyan-output.tar.gz")
		c.Header("Content-Transfer-Encoding", "binary")
		c.Header("Expires", "0")
		c.Header("Cache-Control", "must-revalidate")
		c.Header("Pragma", "public")
		c.DataFromReader(http.StatusOK, -1, "application/x-gzip", pr, nil)
	})

	_ = r.Run(":9000")
}
